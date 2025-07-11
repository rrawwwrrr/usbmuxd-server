package socket

import (
	"net"

	"github.com/sirupsen/logrus"
)

type logWriter struct {
	prefix string
	from   net.Addr
	to     net.Addr
	buffer []byte
	maxLen int
}

const maxLogBytes = 128 // максимум байт для отображения в логе

func (lw *logWriter) Write(p []byte) (n int, err error) {
	// Сохраняем часть данных в буфер
	toPrint := append(lw.buffer, p...)

	// Ограничиваем длину
	if len(toPrint) > maxLogBytes {
		toPrint = toPrint[:maxLogBytes]
	}

	// Логируем как debug
	logrus.WithFields(logrus.Fields{
		"from":   lw.from,
		"to":     lw.to,
		"bytes":  len(p),
		"sample": sanitize(string(toPrint)),
	}).Debug(lw.prefix + "Данные переданы")

	// Сохраняем остаток для следующего вызова
	lw.buffer = append(lw.buffer, p...)
	if len(lw.buffer) > maxLogBytes {
		lw.buffer = lw.buffer[len(lw.buffer)-maxLogBytes:]
	}

	return len(p), nil
}
func sanitize(s string) string {
	res := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= 32 && r <= 126 { // ASCII printable
			res = append(res, r)
		} else {
			res = append(res, '.')
		}
	}
	return string(res)
}
