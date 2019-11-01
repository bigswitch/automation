#!/usr/bin python
import time, os, sys
from controller_bcf import Controller
from pprint import pprint
from lag_manager_config import config
from terminaltables import AsciiTable

controller = Controller(config['controller'], config['access-token'])
prev_member_stats = {}

def clear_counters():
    """ clears counters of members of specified interface groups - NOT USED """
    interface_groups = controller.interface_groups(action='get')
    for interface_group in interface_groups:
        if not interface_group['name'] in config['interface-groups']:
            continue
        if not 'interface' in interface_group:
            continue
        for member in interface_group['interface']:
            interface = member['interface-name']
            switch = member['switch-name']
            switch_dpid = controller.switch_dpid(switch)
            controller.interface_stats(interface, switch_dpid, action='clear')

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
            #print switch_dpid, op_state
            member_counters = {}
            if op_state == 'up':
                stats = controller.interface_stats(interface, switch_dpid)
                #print 'STATS:', stats, ' interface ', switch, interface
                counters = stats['interface'][0]['counter']
                member_counters = {counter: counters[counter] for counter in config['counters']}
            member_stats[(interface_group_name, switch, interface)]  = member_counters, op_state
    return member_stats

def interface_group_monitor():
    count = 0
    # initialize
    global prev_member_stats
    prev_member_stats = get_member_stats()
    
    while True:
        headers = ['intf-grp','switch','intf',]
        for counter in config['counters']:
            headers.append(counter+' \nthreshold='+str(config['counters'][counter]))
        headers += ['state', 'action']
        table = [headers]

        member_stats = get_member_stats()
        for interface_group, switch, interface in member_stats:
            member_counters, op_state = member_stats[(interface_group, switch, interface)]
            prev_member_counters, prev_op_state = prev_member_stats[(interface_group, switch, interface)]

            # write counter delta and op_state to table row
            row = [interface_group, switch, interface, ]
            deltas = {}
            for counter in config['counters']:
                if member_counters and counter in prev_member_counters:
                    delta = 0
                    if prev_member_counters[counter] > member_counters[counter]:
                        delta = member_counters[counter] + (sys.maxint - prev_member_counters[counter])
                    else:
                        delta = member_counters[counter] - prev_member_counters[counter]
                    deltas[counter] = delta
                else:
                    delta = 'N/A'
                row.append(delta)
            row.append(op_state)

            # interface action
            if not member_counters:
                    row.append('None')
            else:
                shutdown = False
                for counter in config['counters']:
                    threshold = config['counters'][counter]
                    if deltas.get(counter, 0) > threshold:
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
        prev_member_stats = member_stats
        time.sleep(config['sampling-interval'])

if __name__ == '__main__':
    interface_group_monitor()
