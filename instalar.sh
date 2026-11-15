#!/bin/bash
# Despliega el servidor como servicio systemd con un usuario sin privilegios.
# Ejecutar como root desde la carpeta del proyecto:  sudo bash instalar.sh
set -e

id proyectoso >/dev/null 2>&1 || useradd -r -s /usr/sbin/nologin proyectoso

install -m 755 bin/servidorOper /usr/local/bin/servidorOper
install -m 640 -o root -g proyectoso conf/configProyecto.conf /etc/configProyecto.conf
install -m 640 -o root -g proyectoso conf/usuarios.txt /etc/usuariosProyecto.txt
install -m 644 systemd/proyecto.service /etc/systemd/system/proyecto.service

systemctl daemon-reload
systemctl enable --now proyecto
systemctl status proyecto --no-pager
