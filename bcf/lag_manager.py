#!/usr/bin python
import time, os, sys
from controller_bcf import Controller
from pprint import pprint
from lag_manager_config import config
from terminaltables import AsciiTable

controller = Controller(config['controller'], config['access-token'])

def get_member_stats():
    member_stats = {}
    interface_groups = controller.interface_groups(action='get')
    #pprint( interface_groups )
    for interface_group in interface_groups:
        if not interface_group['name'] in config['interface-groups']:
            continue
        if not 'interface' in interface_group:
            continue
        interface_group_name = interface_group['name']
        for member in interface_group['interface']:
            interface = member['interface-name']
            op_state = member['op-state']
            switch = member['switch-name']
            switch_dpid = controller.switch_dpid(switch)
            member_counters = {}
            if op_state == 'up':
                stats = controller.interface_stats(interface, switch_dpid)
                counters = stats['interface'][0]['counter']
                member_counters = {counter: counters[counter] for counter in config['counters']}
            member_stats[(interface_group_name, switch, interface)]  = member_counters, op_state
    return member_stats

def interface_group_monitor():
    count = 0
    while True:
        headers = ['intf-grp','switch','intf',]
        for counter in config['counters']:
            headers.append(counter+'/'+str(config['counters'][counter]))
        headers += ['state', 'action']
        table = [headers]

        member_stats = get_member_stats()
        for interface_group, switch, interface in member_stats:
            row = [interface_group, switch, interface, ]
            member_counters, op_state = member_stats[(interface_group, switch, interface)]
            for counter in config['counters']:
                if member_counters:
                    row.append(member_counters[counter])
                else:
                    row.append('N/A')
            row.append(op_state)

            if not member_counters:
                    row.append('None')
            else:
                shutdown = False
                for counter in config['counters']:
                    if member_counters[counter] > config['counters'][counter]:
                        shutdown = True
                if shutdown:
                    # shutdown interface
                    result = controller.interface(switch, interface, action='shutdown')
                    row.append('shutdown')
                else:
                    row.append('None')
            table.append(row)
            
        os.system('clear')
        print AsciiTable(table).table
        print 'Interval:', count
        count += 1
        time.sleep(config['period'])

if __name__ == '__main__':
    interface_group_monitor()
