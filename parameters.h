const int L = 0, R = 100;
const int TASK_SIZE = 10;
const int BUFFER_SIZE = 256;
const int TCP_PORT = 8082;
const int TIME_TO_TASK = 5;
const int KICK_MULT = 2;
const int UDP_PORT = 9999;
const int CLIENT_TIMEOUT = 3;

// func = 3x^2 +2x + 3
inline int integral(int x) {
    return x*x*x + x*x + 3*x;
}
