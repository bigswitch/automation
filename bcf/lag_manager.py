#!/usr/bin python
import time
from controller_bcf import Controller
from pprint import pprint


config = {
    'controller': '52.205.221.120',
    'access-token': 'C3ldz4f4tWV6-DrSHsf63CxU18DU8diw',
    'interface-group': 'R1H1',
    'counter': 'rx-crc-error',
    'threshold': 10,
    'period': 1
}

def get_member_stats(controller):
    interface_group = controller.interface_group(config['interface-group'], action='get')
    member_stats = {}
    for member in interface_group['member-interface']:
        interface = member['interface-name']
        switch = member['switch-name']
        switch_dpid = controller.switch_dpid(switch)
        stats = controller.interface_stats(interface, switch_dpid)
        counters = stats['interface'][0]['counter']
        member_stats[(switch, interface)]  = counters[config['counter']]
    return member_stats

def interface_group_monitor(controller):
    while True:
        member_stats = get_member_stats(controller)
        for switch, interface in member_stats:
            if member_stats[(switch, interface)] > config['threshold']:
                # shutdown interface
                controller.interface(switch, interface, action='shutdown')
            else:
                # noshut interface - check multiple times before bringing the interface up/down
                controller.interface(switch, interface, action='no-shutdown')
        time.sleep(config['period'])

if __name__ == '__main__':
    controller = Controller(config['controller'],
                            config['access-token'])
    interface_group_monitor(controller)
