#!/bin/bash

install --mode=755 --owner=root --group=root --directory /var/log/ecrond
systemctl daemon-reload

#END