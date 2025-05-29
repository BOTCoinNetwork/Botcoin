.. _botcoin_systemd_rst:

Botcoin Systemd Service
----------------------

Here is an example service file defining a ``systemd`` service for botcoin:

``cat /etc/systemd/system/botcoin.service``:

.. code::

[Unit]
Description=Botcoin Node Service
After=network.target
Requires=network-online.target
StartLimitIntervalSec=120

[Service]
Type=simple
User=root
NoNewPrivileges=yes
PrivateTmp=yes
PrivateDevices=yes
DevicePolicy=closed
ProtectControlGroups=yes
ProtectKernelModules=yes
ProtectKernelTunables=yes
# RestrictAddressFamilies=AF_INET AF_INET6
RestrictRealtime=yes
RestrictNamespaces=yes
MemoryDenyWriteExecute=yes
Restart=on-failure
RestartSec=5s
LimitNOFILE=32768
WorkingDirectory=/www/wwwroot/Botcoin
StandardOutput=file:/www/wwwroot/Botcoin/cmd/botcoin/out.log
StandardError=file:/www/wwwroot/Botcoin/cmd/botcoin/info.log
ExecStart=/www/wwwroot/Botcoin/cmd/botcoin/botcoin run -d /root/.monet/monetd-data
ReadWritePaths=/root/.monet

[Install]
WantedBy=multi-user.target

It is fairly locked down and prevents from writing outside of 
``/www/wwwroot/Botcoin/data``.

Note that this requires ``botcoin`` to be installed in ``/opt/botcoin/bin`` and for
the configuration to have been initialised in ``/www/wwwroot/Botcoin/data``. Here, we run 
the service as the ``root`` user, which should have enough permissions in those
directories.

You can then use ``systemctl`` and ``journalctl`` to start, stop, and monitor
the botcoin daemon:

.. code:: 
# 0. permissions
sudo chown -R root:root /www/wwwroot/Botcoin
sudo chmod -R 755 /www/wwwroot/Botcoin

# 1. auto start
sudo systemctl daemon-reload
sudo systemctl enable botcoin
sudo systemctl start botcoin

# 2. check status
sudo systemctl status botcoin   
sudo journalctl -u botcoin -f  

# 3. stop service 
sudo systemctl restart botcoin 
sudo systemctl stop botcoin 
sudo systemctl disable botcoin 

