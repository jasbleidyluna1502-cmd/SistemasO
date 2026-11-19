#!/bin/bash
set -e

id proyectoso >/dev/null 2>&1 || useradd -r -s /usr/sbin/nologin proyectoso

install -m 755 servidorOper /usr/local/bin/servidorOper
install -m 640 -o root -g proyectoso configProyecto.conf /etc/configProyecto.conf
install -m 640 -o root -g proyectoso usuarios.txt /etc/usuariosProyecto.txt
install -m 644 proyecto.service /etc/systemd/system/proyecto.service

systemctl daemon-reload
systemctl enable --now proyecto
systemctl restart proyecto
systemctl status proyecto --no-pager
