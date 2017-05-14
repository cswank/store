# store
A web app for selling stuff.

## notes to be orgainized when I'm really bored:

### Allow a non-root user to start the binary and listen on privileged ports

    $ sudo mv ~/store /usr/local/bin/store; sudo setcap CAP_NET_BIND_SERVICE=+eip /usr/local/bin/store
    $ sudo systemctl restart store.service
