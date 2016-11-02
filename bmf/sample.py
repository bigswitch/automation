#!/usr/bin/env python

from out_of_band import Controller

def main():
    with Controller('10.2.19.102', 'admin', 'bsn') as controller:
        switch_dpid = '00:00:70:72:cf:c1:7c:7b'
        tunnel_name = 'tunnel-vmware0'
        
        tunnel_specs = {
            'switch_dpid': switch_dpid, # Filter-1
            'tunnel_name': tunnel_name,
            'destination_ip': '10.2.19.125',
            'source_ip': '10.10.9.242', # switch uses this IP address for the tunnel interface
            'mask': '255.255.255.240',
            'gateway_ip': '10.10.9.241',
            'vpn_key': 0,
            'encap_type': 'gre',
            'interface': 'ethernet41', # tunnel interface connected to mgmt switch
            'direction': 'receive-only',
            'loopback_interface': ''
        }
        controller.create_tunnel(tunnel_specs, dry_run=False)

        controller.configure_bigtap_interface_role(switch_dpid, tunnel_name, tunnel_name+'-Filter', 'filter')

        policy_specs = {
            'name': 'vmware0',
            'action': 'forward',
            'priority': 100,
            'duration': 0,
            'start_time': 0,
            'delivery_packet_count': 0,
            'interfaces': {
                tunnel_name+'-Filter': 'filter',
                'Bro-IDS-H1-eth3': 'delivery'
            },
            'rules': [
                {"any-traffic": True, "sequence": 1}
            ]
        }        
        controller.add_policy(policy_specs)


if __name__ == '__main__':
    main()