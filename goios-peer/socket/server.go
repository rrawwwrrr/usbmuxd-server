package socket

import (
	"bufio"
	"io"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
)

var destinations = map[string]string{
	"usbmuxd": "UNIX:/var/run/usbmuxd",
	"forward": "TCP:127.0.0.1:7777",
}

func init() {
	// Установи уровень логирования (Debug, Info, Error и т.д.)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
}

// isClosedError проверяет, является ли ошибка "use of closed network connection"
func isClosedError(err error) bool {
	if err == nil {
		return false
	}
	opErr, ok := err.(*net.OpError)
	return ok && (opErr.Err.Error() == "use of closed network connection" || opErr.Err.Error() == "connection reset by peer")
}

// proxy двусторонне проксирует данные между двумя соединениями
func proxy(a, b net.Conn) {
	log.WithFields(log.Fields{
		"from": a.RemoteAddr(),
		"to":   b.RemoteAddr(),
	}).Info("Начало проксирования")

	go func() {
		defer func() {
			a.Close()
			b.Close()
		}()
		if _, err := a.Write(nil); err != nil {
			log.WithError(err).Warn("Соединение A уже закрыто")
			return
		}
		_, err := io.Copy(a, b)
		if err != nil && !isClosedError(err) {
			log.WithError(err).WithFields(log.Fields{
				"source": b.RemoteAddr(),
				"dest":   a.RemoteAddr(),
			}).Error("Ошибка при копировании A<-B")
		}
	}()

	go func() {
		defer func() {
			a.Close()
			b.Close()
		}()
		if _, err := b.Write(nil); err != nil {
			log.WithError(err).Warn("Соединение B уже закрыто")
			return
		}
		_, err := io.Copy(b, a)
		if err != nil && !isClosedError(err) {
			log.WithError(err).WithFields(log.Fields{
				"source": a.RemoteAddr(),
				"dest":   b.RemoteAddr(),
			}).Error("Ошибка при копировании B<-A")
		}
	}()
}

// handleConn обрабатывает входящее соединение
func handleConn(tcpConn net.Conn) {
	defer tcpConn.Close()

	clientIP := tcpConn.RemoteAddr().String()
	log.WithField("client", clientIP).Debug("Новое подключение")

	reader := bufio.NewReader(tcpConn)
	line, err := reader.ReadString('\n')
	if err != nil {
		log.WithError(err).WithField("client", clientIP).Error("Ошибка чтения handshake")
		tcpConn.Write([]byte("Handshake read error\n"))
		return
	}

	key := strings.TrimSpace(line)
	dst, ok := destinations[key]
	if !ok {
		msg := "Unknown destination key\n"
		log.WithFields(log.Fields{
			"client": clientIP,
			"key":    key,
		}).Warn("Неизвестный ключ маршрута")
		tcpConn.Write([]byte(msg))
		return
	}

	parts := strings.SplitN(dst, ":", 2)
	if len(parts) != 2 {
		msg := "Invalid destination format\n"
		log.WithField("dst", dst).Warn("Некорректный формат назначения")
		tcpConn.Write([]byte(msg))
		return
	}

	var remoteConn net.Conn
	proto, addr := parts[0], parts[1]

	switch proto {
	case "UNIX":
		remoteConn, err = net.Dial("unix", addr)
	case "TCP":
		remoteConn, err = net.Dial("tcp", addr)
	default:
		msg := "Unknown protocol\n"
		log.WithField("proto", proto).Warn("Неизвестный протокол")
		tcpConn.Write([]byte(msg))
		return
	}

	if err != nil {
		msg := "Failed to connect to destination\n"
		log.WithError(err).WithFields(log.Fields{
			"proto": proto,
			"addr":  addr,
		}).Error("Ошибка подключения к целевому адресу")
		tcpConn.Write([]byte(msg))
		return
	}

	// Отправляем всё, что было прочитано после handshake
	go func() {
		_, err := io.Copy(remoteConn, reader)
		if err != nil && !isClosedError(err) {
			log.WithError(err).Error("Ошибка передачи данных после handshake")
		}
	}()

	proxy(tcpConn, remoteConn)
}

func Start() {
	ln, err := net.Listen("tcp", ":27015")
	if err != nil {
		log.WithError(err).Fatal("Ошибка запуска сервера")
	}
	defer ln.Close()
	log.Info("TCP-прокси слушает на порту :27015")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.WithError(err).Error("Ошибка при принятии соединения")
			continue
		}
		go handleConn(conn)
	}
}
