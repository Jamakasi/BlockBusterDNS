#!/bin/sh

# Create Tayga directories.
mkdir -p ${TAYGA_CONF_DATA_DIR} ${TAYGA_CONF_DIR}

# Configure Tayga
cat >${TAYGA_CONF_DIR}/tayga.conf <<EOF
tun-device nat64
ipv4-addr ${TAYGA_CONF_IPV4_ADDR}
prefix ${TAYGA_CONF_PREFIX}
dynamic-pool ${TAYGA_CONF_DYNAMIC_POOL}
data-dir ${TAYGA_CONF_DATA_DIR}
EOF

# enable forwarding
sysctl net.ipv6.conf.all.forwarding=1 && sysctl -p

# Setup Tayga networking
tayga -c ${TAYGA_CONF_DIR}/tayga.conf --mktun
ip link set nat64 up
ip route add ${TAYGA_CONF_DYNAMIC_POOL} dev nat64
ip route add ${TAYGA_CONF_PREFIX} dev nat64

# add nat
iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE

# Run Tayga
tayga -c ${TAYGA_CONF_DIR}/tayga.conf -d