import sys
import re
import time
import paramiko


def error_exit(descr):
    print("ERROR: %s" % descr)
    sys.exit(1)


def print_content(content):
    print("content:\n%s\n--------------" % content)


# Usage:
#    dev = BsnCli(args.dev, args.user, args.password, debug=args.debug, debug_level=args.debug_level)
#    content = dev.cmd('show version')
#    dev.close()
#
class BsnCli(object):
    def __init__(self, dev, user, password,
                 debug=False, debug_level=None):
        self._prompt = r'[(\*?)(\w+(-?\w+)?\s?@?)?[\-\w+\.:/]+(?:\([^\)]+\))?(:~)?[>#$] ?$'
        self._timeout = 30
        self._dev = dev
        self._user = user
        self._password = password
        self._debug = debug
        self._debug_level = debug_level or 1
        self._channel = None
        self._ssh = None
        self._cli_content = None
        self.connect()

    def connect(self):
        self.debug("SSH connect to '%s' (user:%s, password:%s)"
                   % (self._dev, self._user, self._password), 1)
        try:
            ssh = paramiko.SSHClient()
            ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
            ssh.connect(self._dev, username=self._user, password=self._password)
            self._channel = ssh.invoke_shell()
            self._ssh = ssh
            self.expect()
        except:
            error_exit("Unable to connect. Please check SSH access to %s." % self._dev)

    def expect(self, prompt=None):
        if prompt == None:
            prompt = self._prompt  # default device prompt
        channel_data = str()
        i = 0
        while True:
            if self._channel.recv_ready():
                channel_data += self._channel.recv(9999)
            else:
                if i > self._timeout:  # reached max timeout
                    error_exit("Reached maximum expect timeout (%s seconds)." % self._timeout)
                    break
                else:
                    time.sleep(1.0)
                    i += 1
                    continue

            last_line = channel_data.splitlines()[-1]
            self.debug("last_line: %s" % last_line, 1)
            if re.match(prompt, last_line):
                self.debug("Found matching prompt", 3)
                break
        self.debug("channel_data: %s" % channel_data, 2)
        self._cli_content = channel_data

        # Return output except for first and last lines which contain
        # the command and the device prompt.
        return channel_data[channel_data.find('\n'):channel_data.rfind('\n')].strip()

    def send(self, command):
        self.debug("Sending command: '%s'" % command, 1)
        self._channel.send(command + '\n')

    def cmd(self, command):
        self.send(command)
        return self.expect()

    def debug(self, descr, level=1):
        if level == None:
            level = self._debug_level
        if self._debug and level <= int(self._debug_level):
            print("DEBUG[%s]: %s" % (level, descr))

    def close(self):
        self.debug("Closing SSH handle for device '%s'" % self._dev, 1)
        self._ssh.close()

