import logging

from apicmd import NetAppCommand

from NaServer import NaElement

LOGGER = logging.getLogger(__name__)


class VlanGetCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'NW.VLAN.GET'

    def execute(self, server, cmd_data_json):
        if (
                "parent_name" not in cmd_data_json or
                "vlan_id" not in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                'get vlan request must have parent name'
                + ' and vlan id defined, got: '
                + str(cmd_data_json))

        parent = cmd_data_json["parent_name"]
        vlan_id = cmd_data_json["vlan_id"]
        cmd = "net-vlan-get-iter"

        call = NaElement(cmd)

        qe = NaElement("query")
        qe_vi = NaElement("vlan-info")
        qe_vi.child_add_string("parent-interface", parent)
        qe_vi.child_add_string("vlanid", vlan_id)
        if "node_name" in cmd_data_json:
            qe_vi.child_add_string(
                "node", cmd_data_json["node_name"])
        qe.child_add(qe_vi)
        call.child_add(qe)

        des_attr = NaElement("desired-attributes")
        vi = NaElement("vlan-info")
        vi.child_add_string("interface-name","<interface-name>")
        vi.child_add_string("node","<node>")
        vi.child_add_string("parent-interface","<parent-interface>")
        vi.child_add_string("vlanid","<vlanid>")
        des_attr.child_add(vi)
        call.child_add(des_attr)
        

        resp, err_resp = self._INVOKE_CHECK(
            server, call, 
            cmd + ": " + parent + ":" + vlan_id)
        if err_resp:
            return err_resp

        #LOGGER.debug(resp.sprintf())

        vlan_cnt = self._GET_INT(resp, 'num-records')
        if not vlan_cnt or vlan_cnt < 1:
            # either None or 0 evaluates to False
            return self._CREATE_FAIL_RESPONSE(
                'no vlans or too many found for query: ['
                + str(cmd_data_json) + '] result is: '
                + resp.sprintf())

        if not resp.child_get("attributes-list"):
            return self._CREATE_FAIL_RESPONSE(
                'no vlan data found in: '
                + resp.sprintf())

        vlan_info = resp.child_get("attributes-list").children_get()[0]

        dd = {
            "name": self._GET_STRING(vlan_info, "interface-name"),
            "node_name": self._GET_STRING(vlan_info, "node"),
            "parent_name": self._GET_STRING(vlan_info, "parent-interface"),
            "vlan_id": self._GET_STRING(vlan_info, "vlanid")
        }

        return {
            'success' : True, 'errmsg': '', 'data': dd}

class VlanCreateCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'NW.VLAN.CREATE'

    def execute(self, server, cmd_data_json):
        if (
                "parent_name" not in cmd_data_json or
                "vlan_id" not in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                'create vlan request must have parent name'
                + ' and vlan id defined, got: '
                + str(cmd_data_json))

        parent = cmd_data_json["parent_name"]
        vlan_id = cmd_data_json["vlan_id"]
        cmd = "net-vlan-create"

        call = NaElement(cmd)
        
        vi = NaElement("vlan-info")
        vi.child_add_string("parent-interface", parent)
        vi.child_add_string("vlanid", vlan_id)
        if "node_name" in cmd_data_json:
            vi.child_add_string(
                "node", cmd_data_json["node_name"])
        call.child_add(vi)

        _, err_resp = self._INVOKE_CHECK(
            server, call, 
            cmd + ": " + parent + ":" + vlan_id)
        if err_resp:
            return err_resp

        return self._CREATE_EMPTY_RESPONSE(
            True, "")

class VlanDeleteCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'NW.VLAN.DELETE'

    def execute(self, server, cmd_data_json):
        if (
                "parent_name" not in cmd_data_json or
                "vlan_id" not in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                'delete vlan request must have parent name'
                + ' and vlan id defined, got: '
                + str(cmd_data_json))

        parent = cmd_data_json["parent_name"]
        vlan_id = cmd_data_json["vlan_id"]
        cmd = "net-vlan-delete"

        call = NaElement(cmd)
        
        vi = NaElement("vlan-info")
        vi.child_add_string("parent-interface", parent)
        vi.child_add_string("vlanid", vlan_id)
        if "node_name" in cmd_data_json:
            vi.child_add_string(
                "node", cmd_data_json["node_name"])
        call.child_add(vi)

        _, err_resp = self._INVOKE_CHECK(
            server, call, 
            cmd + ": " + parent + ":" + vlan_id)
        if err_resp:
            return err_resp

        return self._CREATE_EMPTY_RESPONSE(
            True, "")