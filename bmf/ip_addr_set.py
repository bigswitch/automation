#!/usr/bin/env python

from out_of_band import Controller

def main():
    with Controller('10.2.19.102', 'admin', 'password') as controller:
        # create ipv4 address set 
        controller.create_ip_address_set('ipv4_addr_set', 'ipv4')   
        controller.add_ip_to_group('ipv4_addr_set', '10.10.0.1', '255.255.255.255')
        controller.add_ip_to_group('ipv4_addr_set', '10.10.0.0', '255.255.255.240')

        # create ipv6 address set
        # controller.create_ip_address_set('ipv6_addr_set', 'ipv6')

        # assuming ipv6_addr_set is an address set that has already been created and is used in a policy rule
        controller.add_ip_to_group('ipv6_addr_set', '21DA:D3:0:2F3B:2AA:FF:FE28:9C5A', 'ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff')
        controller.add_ip_to_group('ipv6_addr_set', 'ffff:1234::', 'ffc0:0:0:0:0:0:0:0')

        # delete an ip from a group
        controller.delete_ip_from_group('ipv4_addr_set', '10.10.0.0', '255.255.255.240')

        # for an example of how policies are created see https://github.com/bigswitch/sample-scripts/blob/master/bmf/sample.py

if __name__ == '__main__':
    main()
