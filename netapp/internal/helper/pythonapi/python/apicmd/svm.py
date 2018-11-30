import logging

from apicmd import NetAppCommand, NetAppSvmCommand

from NaServer import NaElement

LOGGER = logging.getLogger(__name__)

class SvmGetCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'SVM.GET'

    def execute(self, server, cmd_data_json):
        if (
                "name" not in cmd_data_json and
                "uuid" not in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                'get SVM request must have name'
                + ' or uuid defined, got: '
                + str(cmd_data_json))

        cmd = "vserver-get-iter"
        call = NaElement(cmd)

        qe = NaElement("query")
        qe_vsi = NaElement("vserver-info")
        if "name" in cmd_data_json:
            qe_vsi.child_add_string(
                "vserver-name",
                cmd_data_json["name"])
        elif "uuid" in cmd_data_json:
            qe_vsi.child_add_string(
                "uuid", cmd_data_json["uuid"])
        qe.child_add(qe_vsi)
        call.child_add(qe)

        des_attr = NaElement("desired-attributes")
        vsi = NaElement("vserver-info")

        for prot_attr in ["allowed-protocols", "disallowed-protocols"]:
            alp = NaElement(prot_attr)
            alp.child_add_string("protocol","<protocol>")
            vsi.child_add(alp)

        vsi.child_add_string("ipspace","<ipspace>")
        vsi.child_add_string("is-config-locked-for-changes","<is-config-locked-for-changes>")
        vsi.child_add_string("operational-state","<operational-state>")
        vsi.child_add_string("operational-state-stopped-reason","<operational-state-stopped-reason>")
        vsi.child_add_string("root-volume","<root-volume>")
        vsi.child_add_string("root-volume-aggregate","<root-volume-aggregate>")
        vsi.child_add_string("root-volume-security-style","<root-volume-security-style>")
        vsi.child_add_string("state","<state>")
        vsi.child_add_string("uuid","<uuid>")
        vsi.child_add_string("volume-delete-retention-hours","<volume-delete-retention-hours>")
        vsi.child_add_string("vserver-name","<vserver-name>")

        des_attr.child_add(vsi)
        call.child_add(des_attr)

        resp, err_resp = self._INVOKE_CHECK(
            server, call, 
            cmd + ": " + str(cmd_data_json))
        if err_resp:
            return err_resp

        #LOGGER.debug(resp.sprintf())

        svm_cnt = self._GET_INT(resp, 'num-records')
        if svm_cnt != 1:
            # not exactly one svm received for query
            return self._CREATE_FAIL_RESPONSE(
                'not exactly one SVM received for: ['
                + str(cmd_data_json) + '] result is: '
                + resp.sprintf())

        if not resp.child_get("attributes-list"):
            return self._CREATE_FAIL_RESPONSE(
                'no svm data found in: '
                + resp.sprintf())

        svm_info = resp.child_get("attributes-list").children_get()[0]

        dd = {
            "name": self._GET_STRING(svm_info, "vserver-name"),
            "uuid": self._GET_STRING(svm_info, "uuid"),
            "ipspace": self._GET_STRING(svm_info, "ipspace"),
            "root_aggr": self._GET_STRING(svm_info, "root-volume-aggregate"),
            "root_sec_style": self._GET_STRING(svm_info, "root-volume-security-style"),
            "root_name": self._GET_STRING(svm_info, "root-volume"),
            "root_retent": self._GET_STRING(svm_info, "volume-delete-retention-hours"),
            
            "locked": self._GET_BOOL(svm_info, "is-config-locked-for-changes"),
            "svm_state": self._GET_STRING(svm_info, "state"),
            
            "proto_enabled": self._GET_CONTENT_LIST(svm_info, "allowed-protocols"),
            "proto_inactive": self._GET_CONTENT_LIST(svm_info, "disallowed-protocols")
        }

        dd["oper_state"] = self._GET_STRING(svm_info, "operational-state")
        if self._GET_STRING(svm_info, "operational-state-stopped-reason"):
            dd["oper_state"] += " caused by: " + self._GET_STRING(
                        svm_info, "operational-state-stopped-reason")

        return {
            'success' : True, 'errmsg': '', 'data': dd}

class SvmCreateCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'SVM.CREATE'

    def execute(self, server, cmd_data_json):
        if (
                "name" not in cmd_data_json or
                "ipspace" not in cmd_data_json or
                "root_aggr" not in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                'create SVM request must have name, '
                + ' ipspace and root aggregate defined, got: '
                + str(cmd_data_json))

        name = cmd_data_json['name']
        ipspace = cmd_data_json['ipspace']

        cmd = "vserver-create-async"
        call = NaElement(cmd)

        call.child_add_string("vserver-name", name)
        call.child_add_string("ipspace", ipspace)
        call.child_add_string(
            "root-volume-aggregate",
            cmd_data_json['root_aggr'])

        if 'root_name' in cmd_data_json:
            call.child_add_string(
                "root-volume", cmd_data_json['root_name'])

        if 'root_sec_style' in cmd_data_json:
            call.child_add_string(
                "root-volume-security-style",
                cmd_data_json['root_sec_style'])

        call.child_add_string("comment", "created by Terraform")

        resp, err_resp = self._INVOKE_CHECK(
            server, call, 
            cmd + ": " + name + " ["
            + ipspace + "]")

        #LOGGER.debug(resp.sprintf())
        
        if err_resp:
            return err_resp

        if (
                not resp.child_get("result") or
                not resp.child_get(
                    'result').child_get('vserver-info')):
            return self._CREATE_FAIL_RESPONSE(
                'no svm data found in: '
                + resp.sprintf())

        svm_info = resp.child_get('result').child_get('vserver-info')

        dd = {
            "name": self._GET_STRING(svm_info, "vserver-name"),
            "uuid": self._GET_STRING(svm_info, "uuid"),
            "ipspace": self._GET_STRING(svm_info, "ipspace"),
            "root_aggr": self._GET_STRING(svm_info, "root-volume-aggregate"),
            "root_sec_style": self._GET_STRING(svm_info, "root-volume-security-style"),
            "root_name": self._GET_STRING(svm_info, "root-volume"),
            "root_retent": self._GET_STRING(svm_info, "volume-delete-retention-hours"),
            
            "locked": self._GET_BOOL(svm_info, "is-config-locked-for-changes"),
            "svm_state": self._GET_STRING(svm_info, "state"),
            
            "proto_enabled": self._GET_CONTENT_LIST(svm_info, "allowed-protocols"),
            "proto_inactive": self._GET_CONTENT_LIST(svm_info, "disallowed-protocols"),

            "status": self._GET_STRING(resp, "result-status")
        }

        if resp.child_get("result-jobid"):
            dd["jobid"] = self._GET_INT(resp, "result-jobid")

        if resp.child_get('result-error-code'):
            dd["errno"] = self._GET_INT(resp, "result-error-code")

        if resp.child_get('result-error-message'):
            dd["errmsg"] = self._GET_STRING(resp, "result-error-message")

        return {
            'success' : True, 'errmsg': '', 'data': dd}

class SvmDeleteCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'SVM.DELETE'

    def execute(self, server, cmd_data_json):
        if (
                "name" not in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                'delete SVM request must have name, '
                + 'defined, got: '
                + str(cmd_data_json))

        name = cmd_data_json['name']

        cmd = "vserver-destroy-async"
        call = NaElement(cmd)

        call.child_add_string("vserver-name", name)
        resp, err_resp = self._INVOKE_CHECK(
            server, call, 
            cmd + ": " + name)

        #LOGGER.debug(resp.sprintf())
        
        if err_resp:
            return err_resp

        dd = {
            "status": self._GET_STRING(resp, "result-status"),
        }

        if resp.child_get("result-jobid"):
            dd["jobid"] = self._GET_INT(resp, "result-jobid")

        if resp.child_get('result-error-code'):
            dd["errno"] = self._GET_INT(resp, "result-error-code")

        if resp.child_get('result-error-message'):
            dd["errmsg"] = self._GET_STRING(resp, "result-error-message")


        return {
            'success' : True, 'errmsg': '', 'data': dd}

class SvmSimpleCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'SVM.SIMPLE.CMD'

    @classmethod
    def _get_simple_cmd(cls):
        raise NotImplementedError('must be implemented by subclass')    

    def execute(self, server, cmd_data_json):
        if (
                "name" not in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                self._get_simple_cmd()
                + ' SVM request must have name '
                + 'defined, got: '
                + str(cmd_data_json))

        name = cmd_data_json['name']

        cmd = "vserver-" + self._get_simple_cmd()
        call = NaElement(cmd)

        call.child_add_string("vserver-name", name)
        # force will only be present if set otherwise ommitted
        if 'force' in cmd_data_json:
            call.child_add_string('force', True)

        resp, err_resp = self._INVOKE_CHECK(
            server, call, 
            cmd + ": " + name)

        #LOGGER.debug(resp.sprintf())
        
        if err_resp:
            return err_resp

        return self._CREATE_EMPTY_RESPONSE(True, "")

class SvmStartCommand(SvmSimpleCommand):

    @classmethod
    def get_name(cls):
        return 'SVM.START'

    @classmethod
    def _get_simple_cmd(cls):
        return 'start'

class SvmStopCommand(SvmSimpleCommand):

    @classmethod
    def get_name(cls):
        return 'SVM.STOP'

    @classmethod
    def _get_simple_cmd(cls):
        return 'stop'

class SvmUnlockCommand(SvmSimpleCommand):

    @classmethod
    def get_name(cls):
        return 'SVM.UNLOCK'

    @classmethod
    def _get_simple_cmd(cls):
        return 'unlock'

class SvmRenameCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'SVM.RENAME'

    def execute(self, server, cmd_data_json):
        if (
                "name" not in cmd_data_json or
                "new_name" not in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                'SVM rename request must have '
                + 'name and new_name defined, got: '
                + str(cmd_data_json))

        name = cmd_data_json['name']
        new_name = cmd_data_json['new_name']

        cmd = "vserver-rename"
        call = NaElement(cmd)

        call.child_add_string("vserver-name", name)
        call.child_add_string("new-name", new_name)

        resp, err_resp = self._INVOKE_CHECK(
            server, call, 
            cmd + ": " + name + ' --> ' + new_name)

        #LOGGER.debug(resp.sprintf())
        
        if err_resp:
            return err_resp

        return self._CREATE_EMPTY_RESPONSE(True, "")

class SvmVolumeSimpleCommand(NetAppSvmCommand):

    @classmethod
    def get_name(cls):
        return 'SVM.VOL.SIMPLECMD'

    @classmethod
    def _get_simple_cmd(cls):
        raise NotImplementedError('must be implemented by subclass')    

    def svm_execute(self, svm, cmd_data_json):
        if (
                "name" not in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                'volume [' + self._get_simple_cmd()
                + '] request must have name '
                + 'defined, got: '
                + str(cmd_data_json))

        name = cmd_data_json['name']

        cmd = "volume-" + self._get_simple_cmd()
        call = NaElement(cmd)

        call.child_add_string("name", name)

        resp, err_resp = self._INVOKE_CHECK(
            svm, call, 
            cmd + ": " + name)

        #LOGGER.debug(resp.sprintf())
        
        if err_resp:
            return err_resp

        return self._CREATE_EMPTY_RESPONSE(True, "")

class SvmVolumeOnlineCommand(SvmVolumeSimpleCommand):

    @classmethod
    def get_name(cls):
        return 'SVM.VOL.ONLINE'

    @classmethod
    def _get_simple_cmd(cls):
        return 'online'

class SvmVolumeOfflineCommand(SvmVolumeSimpleCommand):

    @classmethod
    def get_name(cls):
        return 'SVM.VOL.OFFLINE'

    @classmethod
    def _get_simple_cmd(cls):
        return 'offline'

class SvmVolumeRestrictCommand(SvmVolumeSimpleCommand):

    @classmethod
    def get_name(cls):
        return 'SVM.VOL.RESTRICT'

    @classmethod
    def _get_simple_cmd(cls):
        return 'restrict'

class SvmVolumeDeleteCommand(SvmVolumeSimpleCommand):

    @classmethod
    def get_name(cls):
        return 'SVM.VOL.DELETE'

    @classmethod
    def _get_simple_cmd(cls):
        return 'destroy'

class SvmVolumeSizeCommand(NetAppSvmCommand):

    @classmethod
    def get_name(cls):
        return 'SVM.VOL.SIZE'

    def svm_execute(self, svm, cmd_data_json):
        if (
                "name" not in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                'SVM volume size request must have name '
                + 'defined, got: '
                + str(cmd_data_json))

        name = cmd_data_json['name']

        cmd = "volume-size"
        call = NaElement(cmd)

        call.child_add_string("volume", name)
        if 'size' in cmd_data_json:
            call.child_add_string(
                'new-size', cmd_data_json['size'])

        resp, err_resp = self._INVOKE_CHECK(
            svm, call, 
            cmd + ": " + name)

        #LOGGER.debug(resp.sprintf())
        
        if err_resp:
            return err_resp

        dd = {
            "size": self._GET_STRING(resp, "volume-size"),
        }

        # if resp.child_get("result-jobid"):
        #     dd["jobid"] = self._GET_INT(resp, "result-jobid")

        # if resp.child_get('result-error-code'):
        #     dd["errno"] = self._GET_INT(resp, "result-error-code")

        # if resp.child_get('result-error-message'):
        #     dd["errmsg"] = self._GET_STRING(resp, "result-error-message")

        return {
            'success' : True, 'errmsg': '', 'data': dd}