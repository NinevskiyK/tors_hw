#include "parameters.h"

#include <chrono>
#include <iostream>
#include <stdexcept>
#include <string>
#include <sys/epoll.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <thread>
#include <unistd.h>
#include <netinet/tcp.h>
#include <string.h>
#include <netdb.h>


#define BUFFER_SIZE 256

int sock;
int udp_sock;

sockaddr_in connect_to_server() {
    std::cout << "start connecting..." << std::endl;

    udp_sock = socket(AF_INET, SOCK_DGRAM, 0);
    if (udp_sock == -1) {
        perror("socket");
        throw std::runtime_error("socket");
    }
    int yes = 1;
    if (setsockopt(udp_sock, SOL_SOCKET, SO_REUSEADDR, &yes, sizeof(int)) < 0) {
        perror("setsockopt");
        throw std::runtime_error("socket0");
    }

    struct sockaddr_in udp_addr;
    memset(&udp_addr, 0, sizeof(udp_addr));
    udp_addr.sin_family = AF_INET;
    udp_addr.sin_addr.s_addr = htonl(INADDR_ANY);
    udp_addr.sin_port = htons(UDP_PORT);

    if (bind(udp_sock, (struct sockaddr*)&udp_addr, sizeof(udp_addr)) == -1) {
        perror("bind");
        throw std::runtime_error("bind");
    }

    char udp_buffer[BUFFER_SIZE];
    socklen_t udp_addr_len = sizeof(udp_addr);
    ssize_t udp_bytes_recv = recvfrom(udp_sock, udp_buffer, BUFFER_SIZE - 1, 0, (struct sockaddr*)&udp_addr, &udp_addr_len);
    if (udp_bytes_recv == -1) {
        perror("recvfrom");
        throw std::runtime_error("recvfrom");
    }

    udp_buffer[udp_bytes_recv] = '\0';

    std::cout << "From coordinator: [" << inet_ntoa(udp_addr.sin_addr) << ":" << ntohs(udp_addr.sin_port) << "]: " << udp_buffer << std::endl;

    struct sockaddr_in tcp_addr;
    memset(&tcp_addr, 0, sizeof(tcp_addr));
    tcp_addr.sin_family = AF_INET;
    tcp_addr.sin_addr = udp_addr.sin_addr;
    tcp_addr.sin_port = htons(TCP_PORT);
    return tcp_addr;
}

std::string recv_timeout(int sock) {
    char tcp_buffer[BUFFER_SIZE];

    int epollfd = epoll_create1(0);
    if (epollfd == -1) {
        perror("epoll_create1");
        std::runtime_error("Can't create epoll");
    }

    epoll_event ev;
    ev.events = EPOLLIN;
    ev.data.fd = sock;

    if (epoll_ctl(epollfd, EPOLL_CTL_ADD, sock, &ev) == -1) {
        perror("");
        std::runtime_error("Can't add socket");
    }

    if (epoll_wait(sock, &ev, 1, CLIENT_TIMEOUT) != 1) {
        std::runtime_error("Can't get message from coord!");
    }

    ssize_t tcp_bytes_recv = recv(sock, tcp_buffer, BUFFER_SIZE - 1, 0);
    if (tcp_bytes_recv <= 0) {
        perror("recv");
        throw std::runtime_error("recv");
    }
    tcp_buffer[tcp_bytes_recv] = '\0';
    std::cout << "Received TCP message: " << tcp_buffer << std::endl;
    return tcp_buffer;
}

std::string get_message(int sock) {
    std::string message = "";
    message += recv_timeout(sock);
    while (message.back() != '|') {
        message += recv_timeout(sock);
    }
    std::cout << "Got full message: " << message << std::endl;
    return message.substr(0, message.size() - 1);
}

std::string get_ans(std::string req) {
    std::cout << "calc for " << req << std::endl;
    int a = std::stoi(req);
    std::this_thread::sleep_for(std::chrono::seconds(TIME_TO_TASK));
    return req + "&" + std::to_string(integral(a + TASK_SIZE) - integral(a));
}

void send_back(std::string ans, int sock) {
    std::string message = ans + '|';
    std::cout << "Sending back: " << message << std::endl;
    while (!message.empty()) {
        ssize_t bytes_sent = send(sock, message.c_str(), message.size(), 0);
        if (bytes_sent <= 0) {
            perror("send");
            throw std::runtime_error("send");
        }
        message = message.substr(bytes_sent);
    }
}

void work(sockaddr_in tcp_addr) {
    sock = socket(AF_INET, SOCK_STREAM, 0);
    if (sock == -1) {
        perror("socket");
        throw std::runtime_error("socket");
    }

    if (connect(sock, (struct sockaddr*)&tcp_addr, sizeof(tcp_addr)) == -1) {
        perror("connect");
        close(sock);
        throw std::runtime_error("connect");
    }

    while (true) {
        send_back(get_ans(get_message(sock)), sock);
    }
}
int main() {
    while (true) {
        try {
            work(connect_to_server());
        } catch(...) {
            close(sock);
            close(udp_sock);
            continue;
        }
    }
    return 0;
}