#!/usr/bin/python

import os
import sys
import json
import datetime

# Append the correct path for "BsnCli" class (from bsn_cli.py)
sys.path.append('/root/ansible/lib/ansible/module_utils/network/bigswitch')

from bsn_cli import BsnCli
from ansible.module_utils.basic import *


def bsn_cli_show_version():
    module = AnsibleModule(
        argument_spec = dict(
            controller = dict(required=True),
            user = dict(required=False, default=os.getenv('USER')),
            password = dict(required=False, default=os.getenv('NODE_PASSWORD')),
            scpParams = dict(required=True),
            scpPassword = dict(required=True)
        ),
        supports_check_mode = True
    )

    controller = module.params['controller']
    user = module.params['user']
    password = module.params['password']
    scpParams = module.params['scpParams']
    scpPassword = module.params['scpPassword']

    if not controller:
        module.exit_json(
            stderr="Controller IP parameter is not specified.",
            changed=False,
            rc=1
            )
            
    if not user:
        module.exit_json(
            stderr="User parameter is not specified.",
            changed=False,
            rc=1
            )

    if not password:
        module.exit_json(
            stderr="Password parameter is not specified.",
            changed=False,
            rc=1
            )

    if not scpParams:
        module.exit_json(
            stderr="scpParams parameter is not specified.",
            changed=False,
            rc=1
            )

    if not scpPassword:
        module.exit_json(
            stderr="scpPassword parameter is not specified.",
            changed=False,
            rc=1
            )


    dev = BsnCli(controller, user=user, password=password)
    copyCommand = 'copy running-config scp://'+ scpParams
    content = dev.cmd('enable')
    result = dict(content=content)
    content = dev.send(copyCommand)
    time.sleep(5)
    expectPrompt = dev.expect('(?i).+password: ')
    content = dev.cmd(scpPassword)
    result = dict(content=content)
    dev.close()
    module.exit_json(**result)


if __name__ == '__main__':
    bsn_cli_show_version()
