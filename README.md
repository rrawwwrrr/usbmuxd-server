Этот Docker-образ предназначен для сборки и запуска всех основных библиотек проекта libimobiledevice из исходного кода в среде Ubuntu 24.04.
Контейнер обеспечивает сервис usbmuxd с пробросом сокета через TCP, что удобно для интеграции с другими сервисами.

Состав
Сборка из исходников:
libplist
libtatsu
libusbmuxd
libimobiledevice
libimobiledevice-glue
usbmuxd
Зависимости: Все необходимые библиотеки для сборки и работы с iOS-устройствами.
Сервис: Запуск usbmuxd и проброс сокета через socat.

Флаг --privileged и монтирование /dev/bus/usb нужны для доступа к USB-устройствам.
Порт 27015 пробрасывается наружу для доступа к usbmuxd через TCP.
Описание скрипта запуска (run.sh)
Устанавливает переменную окружения USBUXMD_SOCKET_ADDRESS (по умолчанию — /var/run/usbmuxd).
Запускает socat для проброса unix-сокета usbmuxd наружу через TCP (27015).
Запускает сервис usbmuxd для работы с iOS-устройствами.
(Опционально) можно использовать утилиты из libimobiledevice, например:
idevicepair pair — спарить устройство
ideviceinfo — получить информацию об устройстве
Переменные окружения
USBUXMD_SOCKET_ADDRESS — путь к unix-сокету usbmuxd (по умолчанию /var/run/usbmuxd)
Пример использования
Запустите контейнер:

docker run --rm --privileged \
  -v /dev/bus/usb:/dev/bus/usb \
  -p 27015:27015 \
  libimobiledevice
