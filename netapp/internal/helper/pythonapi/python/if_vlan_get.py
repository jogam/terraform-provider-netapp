import sys
sys.path.append("<path_to_nmsdk_root>/lib/python/NetApp")
from NaServer import *


s = NaServer("192.168.91.200", 1 , 14)
s.set_server_type("FILER")
s.set_transport_type("HTTPS")
s.set_port(443)
s.set_style("LOGIN")
s.set_admin_user("admin", "<password>")



#  find node names for cluster 
api = NaElement("system-node-get-iter")

xi = NaElement("desired-attributes")
api.child_add(xi)


xi1 = NaElement("node-details-info")
xi.child_add(xi1)

xi1.child_add_string("node","<node>")

xo = s.invoke_elem(api)
if (xo.results_status() == "failed") :
    print ("Error:\n")
    print (xo.sprintf())
    sys.exit (1)

print ("Received:\n")
print (xo.sprintf())


api1 = NaElement("net-port-get-iter")

xi_1 = NaElement("desired-attributes")
api1.child_add(xi_1)


xi_1_1 = NaElement("net-port-info")
xi_1.child_add(xi_1_1)

xi_1_1.child_add_string("broadcast-domain","<broadcast-domain>")
xi_1_1.child_add_string("ipspace","<ipspace>")
xi_1_1.child_add_string("link-status","<link-status>")
xi_1_1.child_add_string("node","<node>")
xi_1_1.child_add_string("port","<port>")
xi_1_1.child_add_string("port-type","<port-type>")
xi_1_1.child_add_string("remote-device-id","<remote-device-id>")
xi_1_1.child_add_string("vlan-id","<vlan-id>")
xi_1_1.child_add_string("vlan-node","<vlan-node>")
xi_1_1.child_add_string("vlan-port","<vlan-port>")

xi_1_2 = NaElement("query")
api1.child_add(xi_1_2)


xi_1_3 = NaElement("net-port-info")
xi_1_2.child_add(xi_1_3)

xi_1_3.child_add_string("port-type","physical")

xo1 = s.invoke_elem(api1)
if (xo1.results_status() == "failed") :
    print ("Error:\n")
    print (xo1.sprintf())
    sys.exit (1)

print ("Received:\n")
print (xo1.sprintf())


api2 = NaElement("net-vlan-get-iter")

xi_2 = NaElement("desired-attributes")
api2.child_add(xi_2)


xi_2_1 = NaElement("vlan-info")
xi_2.child_add(xi_2_1)

xi_2_1.child_add_string("gvrp-enabled","<gvrp-enabled>")
xi_2_1.child_add_string("interface-name","<interface-name>")
xi_2_1.child_add_string("node","<node>")
xi_2_1.child_add_string("parent-interface","<parent-interface>")
xi_2_1.child_add_string("vlanid","<vlanid>")

xi_2_2 = NaElement("query")
api2.child_add(xi_2_2)

xi_2_3 = NaElement("vlan-info")
xi_2_2.child_add(xi_2_3)
xi_2_3.child_add_string("vlanid","22")  # <vlanid>22</vlanid>

xo2 = s.invoke_elem(api2)
if (xo2.results_status() == "failed") :
    print ("Error:\n")
    print (xo2.sprintf())
    sys.exit (1)

print ("Received:\n")
print (xo2.sprintf())

