# test implementation of key store
import os
import time
import logging
from enum import Enum

from multiprocessing import Process, Lock, Event, Value
import multiprocessing.managers as mp_Managers

LOGGER = logging.getLogger(__name__)

class RegistryManager(mp_Managers.BaseManager):
   pass


class RegistryServer(Process):
    __KEY = b'NETAPPapiS3CR3Tk8'
    __PORT = int(os.environ.get('NETAPP_API_CR_PORT', '12342'))
    __HOST = 'localhost'

    def __init__(self):
        super(RegistryServer, self).__init__()
        self.notify = Event()

    def wait_complete(self):
        while not self.notify.is_set():
            self.notify.wait(1)

        self.notify.clear()

    @staticmethod
    def GET_CLIENT_REGISTRY():
        RegistryManager.register('get_datadict')
        RegistryManager.register('get_datalock')
        RegistryManager.register('get_sd_event')
        
        manager = RegistryManager(
            address=(
                  RegistryServer.__HOST,
                  RegistryServer.__PORT),
            authkey=RegistryServer.__KEY)
        manager.connect()

        return ClientRegistry(
            manager.get_datadict(),
            manager.get_datalock(),
            manager.get_sd_event())

    def run(self):

        LOGGER.debug('starting up registry server')

        data_dict = {
            'apistatus': 'TBD',     # the status of the NetApp api
            'grpc': {
                'accessmsg': 'TBD'   # the string to return for connection
            },
            'clientsessionstatus' : {
                # for each client --> ID : status .. [RUNNING, SHUTDOWN]
            }
        }

        data_lock = Lock()
        sd_event = Event()
     
        RegistryManager.register('get_datadict',
            callable=lambda: data_dict, proxytype=mp_Managers.DictProxy)
        RegistryManager.register('get_datalock', 
            callable=lambda: data_lock, proxytype=mp_Managers.AcquirerProxy)
        RegistryManager.register('get_sd_event', 
            callable=lambda: sd_event, proxytype=mp_Managers.EventProxy)

        manager = RegistryManager(
                address=(
                    RegistryServer.__HOST,
                    RegistryServer.__PORT),
                authkey=RegistryServer.__KEY)

        LOGGER.debug('registry server created')

        manager.start()

        LOGGER.debug(
            'registry server started @ %s:%s',
            RegistryServer.__HOST,
            RegistryServer.__PORT)

        # indicate to outside world that start is complete
        self.notify.set()

        # sleep until shutdown is issued
        while not sd_event.is_set():
            sd_event.wait(1)

        LOGGER.debug('registry server shutting down')

        # shutdown the manager
        manager.shutdown()
        manager.join()

        LOGGER.debug('registry server exiting')

        # indicate to outside world that stop is complete
        self.notify.set()


class ApiStatus(Enum):
    running = "RUNNING"
    shutdown = "SHUTDOWN"


class ClientStatus(Enum):
    running = "RUNNING"
    shutdown = "SHUTDOWN"
    err_unreg_nodict = 'ERR-UNREG-NODICT'
    err_unreg_unkown = 'ERR-UNREG-UKOWNCLIENT'
    err_unreg_wrdict = 'ERR-UNREG-WRDICT'
    err_gstat_nodict = 'ERR-GETSTATUS-NODICT'
    err_gstat_unkown = 'ERR-GETSTATUS-UNKOWNCLIENT'

    @classmethod
    def is_error(cls, status):
        return status in [
                ClientStatus.err_unreg_nodict, 
                ClientStatus.err_unreg_unkown, 
                ClientStatus.err_gstat_nodict, 
                ClientStatus.err_gstat_unkown
        ]


class ClientRegistry(object):

    def __init__(self, data_dict, data_lock, sd_event):
        self.data_dict = data_dict
        self.data_lock = data_lock
        self.sd = sd_event

    def __get_value(self, key):
        self.data_lock.acquire()
        value = self.data_dict.get(key, None)
        self.data_lock.release()

        return value

    def __set_value(self, key, value):
        if key not in self.data_dict:
            return False

        self.data_lock.acquire()
        self.data_dict[key] = value
        self.data_lock.release()

        return True
    
    def get_api_status(self):
        return self.__get_value('apistatus')

    def set_api_status(self, status):
        return self.__set_value('apistatus', status)

    def shutdown(self):
        # shutdown all clients
        succ, cdict = self.__get_client_dict()
        if not succ:
            LOGGER.error('cannot trigger client shutdown')
        else:
            for cid in cdict.keys():
                self.set_client_status(
                    cid, ClientStatus.shutdown, clients=cdict)
        
        # don't need to write back clients just wait for dict to empty
        while self.__get_client_count() > 0:
                time.sleep(1)
        
        # set the shutdown event
        self.sd.set()

    def get_grpc_msg(self):
        grpc_dict = self.__get_value('grpc')
        if type(grpc_dict) is not dict:
            LOGGER.error('expected dict for key[grpc], got: %s', grpc_dict)
            return 'no message for you bud...'

        return grpc_dict.get('accessmsg', 'no msg in grpc')

    def set_grpc_msg(self, msg):
        grpc_dict = self.__get_value('grpc')
        if type(grpc_dict) is not dict:
            LOGGER.error('expected dict for key[grpc], got: %s', grpc_dict)
            return False

        grpc_dict['accessmsg'] = msg
        return self.__set_value('grpc', grpc_dict)

    def __get_client_dict(self):
        cdict = self.__get_value('clientsessionstatus')
        if type(cdict) is not dict:
            LOGGER.error('expected dict for key[css], got: %s', cdict)
            return False, {}

        return True, cdict

    def __get_client_count(self):
        cdict = self.__get_value('clientsessionstatus')
        if type(cdict) is not dict:
            LOGGER.error('expected dict for key[css], got: %s', cdict)
            return 0

        return len(cdict)

    def register_client(self, id, status):
        succ, cdict = self.__get_client_dict()
        if not succ:
            LOGGER.error('could not register client ID [%s]', id)
            return succ

        if id in cdict:
            LOGGER.warn('client [%s] already registered...', id)
            
        return self.set_client_status(id, status, cdict)

    def unregister_client(self, id):
        succ, cdict = self.__get_client_dict()
        if not succ:
            LOGGER.error('could not unregister client ID [%s]', id)
            return ClientStatus.err_unreg_nodict

        if id not in cdict:
            LOGGER.warn('trying to unregister unknown client ID [%s]', id)
            return ClientStatus.err_unreg_unkown

        status = cdict.pop(id)
        if not self.__set_value('clientsessionstatus', cdict):
            LOGGER.error(
                'could not write back client dict during unregister of client [%s]', id)
            status = ClientStatus.err_unreg_wrdict

        return status
        
        
    def set_client_status(self, id, status, clients=None):
        cdict = clients

        if not cdict:
            succ, cdict = self.__get_client_dict()
            if not succ:
                LOGGER.error('could not update status for client ID [%s]', id)
                return succ

        cdict[id] = status
        return self.__set_value('clientsessionstatus', cdict)

    def get_client_status(self, id):
        succ, cdict = self.__get_client_dict()
        if not succ:
            LOGGER.error('could not get status for client ID [%s]', id)
            return ClientStatus.err_gstat_nodict

        if id not in cdict:
            LOGGER.warn('trying to get status for unknown client ID [%s]', id)
            return ClientStatus.err_gstat_unkown

        return cdict.get(id)

class CallCounter(object):
    '''
    multiprocessing/threading save counter
    inspired by: https://eli.thegreenplace.net/2012/01/04/shared-counter-with-pythons-multiprocessing
    '''

    def __init__(self, init_alive=True):
        self.call_count = Value('i', 0)
        self.active_calls = Value('i', 0)
        self.init_alive = init_alive
        self.lock = Lock()

    def start_call(self):
        with self.lock:
            self.call_count.value += 1
            self.active_calls.value += 1

    def end_call(self):
        with self.lock:
            self.active_calls.value -= 1

    def get_call_cnt(self, reset=True):
        ccnt = 0
        with self.lock:
            ccnt = self.call_count.value
            if reset:
                self.call_count.value = 0
            
        return ccnt

    def stay_alive(self):
        if self.init_alive:
            # initial stay alive return req'd
            self.init_alive = False
            return True

        if self.get_call_cnt(reset=False) > 0:
            # need to stay alive non-zero call count
            return True

        with self.lock:
            # stay alive on non-ended active calls
            return self.active_calls.value > 0

    def reset(self):
        with self.lock:
            self.call_count.value = 0
            self.active_calls.value = 0
