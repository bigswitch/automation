# Getting Started with Ansible

## Install Ansible
Install Ansible in your system. For example here are the instructions for Debian systems:
```bash
Latest Releases via Apt (Debian)
Debian users may leverage the same source as the Ubuntu PPA.

Add the following line to /etc/apt/sources.list:

deb http://ppa.launchpad.net/ansible/ansible/ubuntu trusty main
Then run these commands:

$ sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 93C4A3FD7BB9C367
$ sudo apt update
$ sudo apt install ansible
```

Refer to the Ansible installation guide for more detailed instructions:
https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html#installation-guide

## Confirm the Ansible Installation
Confirm the version of Ansible (must be >= 2.5):

```
ansible --version
```