import socket


def recv_all(socket: socket.socket, size):
    n = size
    data = b""
    while n > 0:
        new_data = socket.recv(n)
        if not new_data:
            raise ConnectionError("conexión cerrada")
        data += new_data
        n -= len(new_data)
    return data


def send_all(socket: socket.socket, bytes):
    sent = 0
    while sent < len(bytes):
        delta_n = socket.send(bytes[sent:])
        if not delta_n:
            raise ConnectionError("conexión cerrada")
        sent += delta_n
