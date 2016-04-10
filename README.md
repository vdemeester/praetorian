# Praetorian

[![Join the chat at https://gitter.im/vdemeester/praetorian](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/vdemeester/praetorian?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge) 
[![Build Status](https://travis-ci.org/vdemeester/praetorian.svg?branch=add-readme)](https://travis-ci.org/vdemeester/praetorian)
[![Go Report Card](https://goreportcard.com/badge/github.com/vdemeester/praetorian)](https://goreportcard.com/report/github.com/vdemeester/praetorian)
[![License](https://img.shields.io/github/license/vdemeester/praetorian.svg)]()

<img src="http://raw.github.com/vdemeester/praetorian/master/imgs/praetorian.png"
 alt="Praetorian logo" title="The man himself" align="right" />

Praetorian is a command to be used as an ssh command that allow multiple
commands for multiple ssh keys. It is similar to [sshcommand](https://github.com/progrium/sshcommand)
for the ``$HOME/.ssh/authorized_keys`` part, as it uses the same format.

The basic idea is to allow a set of commands for an identity (a.k.a.
an ssh key). Each identities are identified by an alias, a given
name for the public keys.

**Note : This is an alpha software, use at your own risk. There is still
rough edges and it's not as secure as it should be. And it's being re-written in golang.. :D**

[![Flattr this git repo](http://api.flattr.com/button/flattr-badge-large.png)](https://flattr.com/submit/auto?user_id=vdemeester&url=http://github.com/vdemeester/praetorian&title=praetorian&language=&tags=github&category=software) 

## Usage

To setup praetorian, you'll need the ssh public key and that's pretty much it.

    $ cat ~/.id_rsa.pub | ssh user@host praetorian setup myalias

Next you need to edit the configuration file on the remote, see the [next section](#praetorian-configuration).
Let say we add ``ls`` and ``nc`` as allowed commands (nc for allowing ssh gateway via ProxyCommand).
Now you have some commands allowed, let's try it.

    $ ssh user@host ls
    src
    public_html
    $ ssh user@host pwd
    # Nothing, just exit 1
    praetorian-wrapper: Alias gohei Invalid command pwd
    $ ssh user@host nc -w 1 host2 22
    (host2) $

<!--
Now, if the user identified with this ssh key is connecting, it will read the
``$HOME/.ssh/praetorian`` file, on the remote, to look what commands are allowed.
The commands are looked up by the given alias, and you can add commands on the
remote using praetorian command.

    (remote) $ praetorian add myalias rsync # will add rsync to the allowed commands

There's few command still :

    (remote) $ praetorian list myalias      # list the allowed commands for the alias

    (remote) $ praetorian rm myalias rsync  # will remove rsync from the allowed commands
    (remote) $ praetorian unset myalias     # remove the alias (and the keys) from the authorized_keys
-->

## Praetorian configuration

The configuration file is located at ``$HOME/.ssh/praetorian`` and is, for the
moment, a simple shell-like file.

    (remote) $ cat $HOME/.ssh/praetorian
    myalias="command1 command2 command3"
    gohei="nc cowsay"

<!--
## How does it works

- Using ssh ``authorized_keys`` options
- Reading config file and executing the wrapper command

## Troubleshootings

- ssh command to force password (if needed)
- ssh command to force an identity (ssh key)
-->


[![Bitdeli Badge](https://d2weczhvl823v0.cloudfront.net/vdemeester/praetorian/trend.png)](https://bitdeli.com/free "Bitdeli Badge")

