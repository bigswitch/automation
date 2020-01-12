# BCF EVPC Examples

## Create BCF EVPCs/Tenants & Segments

#### Use the following ansible script to create BCF EVPCs/Tenants & Segments. You will be able to create following skeleton config configured via the above ansible playbook:

```bash
BCF-CTRL-1# show run tenant 3-tier-app

! tenant
tenant 3-tier-app
  logical-router
    route 0.0.0.0/0 next-hop tenant system
    interface segment app
      ip address 10.0.1.1/24
    interface segment fw-01
      ip address 10.0.5.1/24
    interface segment web
      ip address 10.0.0.1/24
    policy-list Firewall
      10 permit segment-interface web any to tenant 3-tier-app segment app next-hop next-hop-group ServiceNode
      20 permit any to any
  segment app
    member interface-group R2H1 vlan untagged
  segment fw-01
    member interface-group FW-01 vlan untagged
  segment web
    member interface-group R1H1 vlan untagged
    member interface-group R1H2 vlan untagged
BCF-CTRL-1#
```

### Use the following ansible files:
- [bcf.yml](https://github.com/bigswitch/automation/blob/master/ansible/bcf.yml)
- [Modules](https://github.com/bigswitch/automation/tree/master/ansible/modules)

bcf.yml:
```ansible
---
- hosts: cont16
  connection: local
  gather_facts: no
  any_errors_fatal: true

  vars:
    inter_segment_fw_rule:  {"seq": 10,
                             "action": "next-hop",
                             "segment-interface": "web",
                             "dst": {"segment": "app", "tenant": "3-tier-app"},
                             "next-hop": {"next-hop-group": "ServiceNode"}}

    deny_segment_to_tenant:  {"seq": 15,
                              "action": "deny",
                              "segment-interface": "web",
                              "dst": {"segment": "qa", "tenant": "app-test"}}

    permit_any_to_any_rule: {"seq": 20, "action": "permit"}

  tasks:
    - name: bcf tenant
      bcf_tenant:
        name: 3-tier-app
        logical_router_interfaces: {'web': '10.0.0.1/24',
                                    'app':'10.0.1.1/24',
                                    'fw-01':'10.0.5.1/24'}
#        logical_router_system_tenant_interface: {'state':'present'}
        routes: {'0.0.0.0/0':'tenant:system'}
#        next_hop_groups: {'ServiceNode':'10.0.5.2'}
        policy_lists: {'Firewall': ['{{ inter_segment_fw_rule }}',
                                    #'{{ deny_segment_to_tenant }}',
                                    '{{ permit_any_to_any_rule }}']}
#        inbound_policy: Firewall
        controller: '{{ inventory_hostname }}'
        state: present
        access_token: cJHUJHGCGvXjOH_HaYJ3QgarRPVcv--5

    - name: bcf segment
      bcf_segment:
        name: web
        interface_groups: [ 'R1H1:untagged', 'R1H2:untagged' ]
        tenant: 3-tier-app
        controller: '{{ inventory_hostname }}'
        state: present
        access_token: cJHUJHGCGvXjOH_HaYJ3QgarRPVcv--5

    - name: bcf segment
      bcf_segment:
        name: app
        interface_groups: ['R2H1:untagged']
        tenant: 3-tier-app
        controller: '{{ inventory_hostname }}'
        state: present
        access_token: cJHUJHGCGvXjOH_HaYJ3QgarRPVcv--5

    - name: bcf segment
      bcf_segment:
        name: fw-01
        interface_groups: ['FW-01:untagged']
        tenant: 3-tier-app
        controller: '{{ inventory_hostname }}'
        state: present
        access_token: cJHUJHGCGvXjOH_HaYJ3QgarRPVcv--5
```

