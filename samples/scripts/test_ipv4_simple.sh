# add namespaces
ip netns add ns1
ip netns add ns2
ip netns add ns3

# Veth pair between ns1 and bridge
ip link add veth-ns1 type veth peer name veth-bridge-ns1
ip link set veth-ns1 netns ns1

# Veth pair between ns2 and bridge
ip link add veth-ns2 type veth peer name veth-bridge-ns2
ip link set veth-ns2 netns ns2

# Veth pair between ns3 and bridge
ip link add veth-ns3 type veth peer name veth-bridge-ns3
ip link set veth-ns3 netns ns3

# Direct veth pair between ns1 and ns3
ip link add veth-ns1-ns3 type veth peer name veth-ns3-ns1
ip link set veth-ns1-ns3 netns ns1
ip link set veth-ns3-ns1 netns ns3

# Create the bridge in the default namespace
ip link add name bridge0 type bridge
ip link add name bridge1 type bridge

# Set bridge up
ip link set bridge0 up
ip link set bridge1 up

# Attach veth interfaces to the bridge
ip link set veth-bridge-ns3 master bridge0

ip link set veth-bridge-ns1 master bridge1
ip link set veth-bridge-ns2 master bridge1

# Bring up the veth interfaces in the default namespace
ip link set veth-bridge-ns1 up
ip link set veth-bridge-ns2 up
ip link set veth-bridge-ns3 up


# ns1
ip netns exec ns1 ip addr add 192.168.1.1/24 dev veth-ns1
ip netns exec ns1 ip link set veth-ns1 up

ip netns exec ns1 ip addr add 192.168.2.1/24 dev veth-ns1-ns3
ip netns exec ns1 ip link set veth-ns1-ns3 up

# ns2
ip netns exec ns2 ip addr add 192.168.1.2/24 dev veth-ns2
ip netns exec ns2 ip link set veth-ns2 up

# ns3
ip netns exec ns3 ip addr add 192.168.1.3/24 dev veth-ns3
ip netns exec ns3 ip link set veth-ns3 up

ip netns exec ns3 ip addr add 192.168.2.2/24 dev veth-ns3-ns1
ip netns exec ns3 ip link set veth-ns3-ns1 up