#include <chrono>
#include <cstdio>
#include <cstring>
#include <iomanip>
#include <iostream>
#include <map>
#include <optional>
#include <ostream>
#include <stack>
#include <stdexcept>
#include <string>
#include <sys/socket.h>
#include <sys/epoll.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <thread>
#include <unistd.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <netinet/tcp.h>
#include <unordered_set>
#include <vector>

#include "parameters.h"

void broadcast_ip() {
    int sock = socket(AF_INET, SOCK_DGRAM, 0);
    if (sock < 0) {
        throw std::runtime_error("Can't create socket!");
    }

    sockaddr_in broadcastAddr;
    broadcastAddr.sin_family = AF_INET;
    broadcastAddr.sin_port = htons(UDP_PORT);
    broadcastAddr.sin_addr.s_addr = htonl(INADDR_BROADCAST);

    int optval = 1;
    if (setsockopt(sock, SOL_SOCKET, SO_BROADCAST, &optval, sizeof(optval))) {
        perror("setsockopt");
        throw std::runtime_error("Can't upgrade to broadcast socket");
    }

    while (true) {
        std::string message = "start working";
        const char* msg = message.c_str();

        ssize_t numBytes = sendto(sock, msg, strlen(msg), 0, (struct sockaddr*)&broadcastAddr, sizeof(broadcastAddr));
        if (numBytes < 0) {
            std::runtime_error("Can't send broadcast message!");
        }
        std::this_thread::sleep_for(std::chrono::seconds(1));
    }

    close(sock);
}

/// EPOLL

const int MAX_EVENTS = 10;
int epollfd = 0;

void add_to_epoll(int sock) {
    epoll_event ev;
    ev.events = EPOLLOUT | EPOLLIN | EPOLLERR | EPOLLHUP;
    ev.data.fd = sock;
    if (epoll_ctl(epollfd, EPOLL_CTL_ADD, sock, &ev) == -1) {
        perror("epoll_ctl");
        throw std::runtime_error("Can't add socket to epoll_ctl");
    }
}

void turn_all(int sock) {
    epoll_event ev;
    ev.events = EPOLLERR | EPOLLHUP;
    ev.data.fd = sock;
    if (epoll_ctl(epollfd, EPOLL_CTL_MOD, sock, &ev) == -1) {
        perror("epoll_ctl");
        throw std::runtime_error("Can't add socket to epoll_ctl");
    }
}

void turn_only_errors(int sock) {
    epoll_event ev;
    ev.events = EPOLLOUT | EPOLLIN | EPOLLERR | EPOLLHUP;
    ev.data.fd = sock;
    if (epoll_ctl(epollfd, EPOLL_CTL_MOD, sock, &ev) == -1) {
        perror("epoll_ctl");
        throw std::runtime_error("Can't add socket to epoll_ctl");
    }
}


int create_epoll(int listen_sock) {
    int epollfd = epoll_create1(0);
    if (epollfd == -1) {
        std::runtime_error("Can't create epoll");
    }
    epoll_event ev;
    ev.events = EPOLLIN;
    ev.data.fd = listen_sock;
    if (epoll_ctl(epollfd, EPOLL_CTL_ADD, listen_sock, &ev) == -1) {
        std::runtime_error("Can't add listen socket");
    }
    return epollfd;
}

/*
 * Инварианты:
 * Каждая непосчитанная таска либа в tasks, либо в worker2task (но не там и там)
 * Как только worker присылает ответ, проверяется, что он и правда считал для этой таски
 * Если воркер долго не отвечает/прервалось соединение, то его задача перекладывается в task, будится один relaxed сокет.
*/
std::map<int, int> worker2task;
std::map<int, std::chrono::time_point<std::chrono::steady_clock>> worker2task_time;
std::map<int, std::string> worker2buffer;
std::map<int, std::string> worker2write_buffer;
std::stack<int> tasks;
std::unordered_set<int> relaxed;

int sum = 0;

std::optional<int> get_task_for_worker(int sock) {
    if (tasks.empty() && worker2task.empty()) {
        std::cout << "ANS: " << sum;
        exit(0);
    }
    if (tasks.empty()) {
        return std::nullopt;
    }
    int task = tasks.top();
    tasks.pop();
    worker2task[sock] = task;
    worker2task_time[sock] = std::chrono::steady_clock::now();
    return task;
}

void give_new_task(int sock);

void readd_task(int task) {
    tasks.push(task);
    if (!relaxed.empty()) {
        turn_all(*relaxed.begin());
        give_new_task(*relaxed.begin());
        relaxed.erase(relaxed.begin());
    }
}

void relax(int sock) {
    relaxed.insert(sock);
    turn_only_errors(sock);
}

void give_new_task(int sock) {
    if (worker2task.count(sock)) {
        worker2task.erase(sock);
    }
    auto task_or_nullopt = get_task_for_worker(sock);
    if (!task_or_nullopt.has_value()) {
        relax(sock);
        return;
    }

    int task = task_or_nullopt.value();
    worker2write_buffer[sock] = std::to_string(task) + '|';
    std::cout << "Message for worker" << sock << ": " << worker2write_buffer[sock] << std::endl;
}


void new_worker(int listen_sock) {
    sockaddr addr;
    socklen_t addrlen;
    int conn_sock = accept4(listen_sock, (struct sockaddr *) &addr, &addrlen, SOCK_NONBLOCK);
    if (conn_sock == -1) {
        perror("accept4");
        return;
    }

    add_to_epoll(conn_sock);
    std::cout << "New worker: " << conn_sock << std::endl;
    give_new_task(conn_sock);
}

int start_listen() {
    int listenfd = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);
    int yes = 1;

    if (listenfd == -1) {
        perror("socket");
        throw std::runtime_error("socket in listen");
    }

    if (setsockopt(listenfd, SOL_SOCKET, SO_REUSEADDR, &yes, sizeof(int)) < 0) {
        perror("setsockopt");
        throw std::runtime_error("socket0");
    }


    int keepAlive = 1;
    if (setsockopt(listenfd, SOL_SOCKET, SO_KEEPALIVE, &keepAlive, sizeof(keepAlive)) == -1) {
        perror("setsockopt");
        throw std::runtime_error("socket1");
    }

    int keepIdle = 2;
    int keepInterval = 2;
    int keepCount = 3;

    if (setsockopt(listenfd, IPPROTO_TCP, TCP_KEEPIDLE, &keepIdle, sizeof(keepIdle)) == -1) {
        perror("setsockopt");
        close(listenfd);
        throw std::runtime_error("socket2");
    }

    if (setsockopt(listenfd, IPPROTO_TCP, TCP_KEEPINTVL, &keepInterval, sizeof(keepInterval)) == -1) {
        perror("setsockopt");
        close(listenfd);
        throw std::runtime_error("socket3");
    }

    if (setsockopt(listenfd, IPPROTO_TCP, TCP_KEEPCNT, &keepCount, sizeof(keepCount)) == -1) {
        perror("setsockopt");
        close(listenfd);
        throw std::runtime_error("socket4");
    }

    struct sockaddr_in serv_addr;
    memset(&serv_addr, 0, sizeof(serv_addr));
    serv_addr.sin_family = AF_INET;
    serv_addr.sin_addr.s_addr = htonl(INADDR_ANY);
    serv_addr.sin_port = htons(TCP_PORT);

    if (bind(listenfd, (struct sockaddr *)&serv_addr, sizeof(serv_addr)) == -1) {
        perror("bind");
        close(listenfd);
        throw std::runtime_error("bind");
    }

    if (listen(listenfd, 10) == -1) {
        perror("listen");
        close(listenfd);
        throw std::runtime_error("listen");
    }

    return listenfd;
}

void close_connection(int sock) {
    std::cout << "Worker is out: " << sock << std::endl;

    close(sock);
    if (worker2task.count(sock)) {
        readd_task(worker2task[sock]);
        worker2task.erase(sock);
    }
    if (worker2buffer.count(sock)) {
        worker2buffer.erase(sock);
    }
    if (worker2write_buffer.count(sock)) {
        worker2write_buffer.erase(sock);
    }
    if (worker2task_time.count(sock)) {
        worker2task_time.erase(sock);
    }
    if (relaxed.count(sock)) {
        relaxed.erase(sock);
    }

    epoll_ctl(epollfd, EPOLL_CTL_DEL, sock, NULL);
}

void kick_inactive() {
    std::vector<int> close;
    for (auto it: worker2task_time) {
        auto duration = std::chrono::duration_cast<std::chrono::seconds>(std::chrono::steady_clock::now() - it.second).count();
        if (duration > TIME_TO_TASK * KICK_MULT) {
            close.push_back(it.first);
        }
    }
    for (auto i: close) {
        close_connection(i);
    }
}


void parse_message(int sock) {
    auto sep = worker2buffer[sock].find('&');
    if (sep == std::string::npos) {
        std::cout << "wrong message: " << worker2buffer[sock] << " from " << sock << std::endl;
        close_connection(sock);
        return;
    }
    std::string ans = worker2buffer[sock].substr(sep + 1, worker2buffer[sock].size() - 1);
    std::string task = worker2buffer[sock].substr(0, sep);
    worker2buffer.erase(sock);
    if (worker2task[sock] != std::stoi(task)) {
        std::cout << "got ans {" << ans << "} from wrong task {" << task << "} from worker {" << sock << "} with task {" << worker2task[sock] << "}" << std::endl;
    } else {
        int res = std::stoi(ans);
        sum += res;
    }
}

void read_from(int sock) {
    char buffer[BUFFER_SIZE];
    ssize_t bytes_received = recv(sock, buffer, BUFFER_SIZE - 1, 0);

    if (bytes_received == -1) {
        if (errno == EAGAIN || errno == EWOULDBLOCK) {
            return;
        } else {
            close_connection(sock);
        }
    } else if (bytes_received == 0) {
        close_connection(sock);
        return;
    }
    buffer[bytes_received] = '\0';
    std::cout << "message " << buffer << " from worker " << sock << std::endl;
    worker2buffer[sock] += buffer;
    if (buffer[bytes_received - 1] == '|') {
        parse_message(sock);
        give_new_task(sock);
    }
}

void write_to(int sock) {
    if (!worker2write_buffer.count(sock)) {
        return;
    }
    ssize_t bytes_sent = send(sock, worker2write_buffer[sock].c_str(), worker2write_buffer[sock].size(), 0);

    if (bytes_sent == -1) {
        if (errno == EAGAIN || errno == EWOULDBLOCK) {
            return;
        } else {
            close_connection(sock);
        }
    }
    worker2write_buffer[sock] = worker2write_buffer[sock].substr(bytes_sent);
    if (worker2write_buffer[sock].size() == 0) {
        worker2write_buffer.erase(sock);
    }
}

void handle_epoll(epoll_event ev, int listen_sock) {
    if (ev.data.fd == listen_sock) {
        new_worker(listen_sock);
    } else {
        if (ev.events & EPOLLIN) {
            read_from(ev.data.fd);
        } else if (ev.events & EPOLLOUT) {
            write_to(ev.data.fd);
        } else {
            close_connection(ev.data.fd);
        }
    }
}

void init_tasks() {
    for (int i = L; i < R; i += TASK_SIZE) {
        tasks.push(i);
    }
}

void init_broadcaster() {
    std::thread broadcaster(broadcast_ip);
    broadcaster.detach();
}

void start_polling() {
    int listen_sock = start_listen();

    epoll_event events[MAX_EVENTS];
    epollfd = create_epoll(listen_sock);

    int nfds;
    for (;;) {
        nfds = epoll_wait(epollfd, events, MAX_EVENTS, 1000);

        for (int i = 0; i < nfds; ++i) {
            handle_epoll(events[i], listen_sock);
        }

        kick_inactive();
    }
}
int main() {
    init_tasks();
    init_broadcaster();
    start_polling();
    return 0;
}