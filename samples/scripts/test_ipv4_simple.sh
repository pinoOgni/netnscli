#!/usr/bin/bash

usage() {
	echo "netnscli usage:"
	echo ""
	echo "$0 [ACTION]"
	echo ""
	echo "ACTIONS"
	echo "		apply		Apply the testbed"
	echo ""
	echo "		delete		delete the testbed"
	echo ""
}

die() {
	echo "Unrecoverable error: $1"
	exit 1
}

apply() {
	# add namespaces
	ip netns add ns1
	ip netns add ns2
	ip netns add ns3
	
	# Create veth pairs
	ip link add veth-ns1 type veth peer name veth-bridge-ns1
	ip link set veth-ns1 netns ns1
	ip netns exec ns1 ip addr add 192.168.1.1/24 dev veth-ns1
	ip netns exec ns1 ip link set veth-ns1 up
	ip link set veth-bridge-ns1 up
	
	ip link add veth-ns2 type veth peer name veth-bridge-ns2
	ip link set veth-ns2 netns ns2
	ip netns exec ns2 ip addr add 192.168.1.2/24 dev veth-ns2
	ip netns exec ns2 ip link set veth-ns2 up
	ip link set veth-bridge-ns2 up
	
	ip link add veth-ns3 type veth peer name veth-bridge-ns3
	ip link set veth-ns3 netns ns3
	ip netns exec ns3 ip addr add 192.168.1.3/24 dev veth-ns3
	ip netns exec ns3 ip link set veth-ns3 up
	ip link set veth-bridge-ns3 up
	
	ip link add veth-ns1-ns3 type veth peer name veth-ns3-ns1
	ip link set veth-ns1-ns3 netns ns1
	ip link set veth-ns3-ns1 netns ns3
	ip netns exec ns1 ip addr add 192.168.2.1/24 dev veth-ns1-ns3
	ip netns exec ns3 ip addr add 192.168.2.2/24 dev veth-ns3-ns1
	ip netns exec ns1 ip link set veth-ns1-ns3 up
	ip netns exec ns3 ip link set veth-ns3-ns1 up
	
	# Create and configure bridges
	ip link add name bridge0 type bridge
	ip link set bridge0 up
	ip link set veth-bridge-ns3 master bridge0
	
	ip link add name bridge1 type bridge
	ip link set bridge1 up
	ip link set veth-bridge-ns1 master bridge1
	ip link set veth-bridge-ns2 master bridge1
	
}

delete() {
	ip link set veth-bridge-ns3 nomaster
	ip link set veth-bridge-ns1 nomaster
	ip link set veth-bridge-ns2 nomaster
	
	ip netns del ns1
	ip netns del ns2
	ip netns del ns3
	
	ip link del bridge0
	ip link del bridge1
}

if [ -z $1 ]; then
	die "You must specify an action between apply and delete"
fi

if [ $1 == "apply" ]; then
	apply
elif [ $1 == "delete" ]; then
	delete
else
	die "$1 is not an existing netnscli action"
fi
