tc\_reader
=========

A persistent SNMP script that exports TC Queue and Class statistics for graphing
(for example to Cacti).

If you ever needed to setup proper QoS configuration on your Linux based router
and then wondered how to monitor the individual Queue / Class statistics or even
better monitor individual users, you probably know the pain. You either ended
being disappointed or implemented your own solution.

If you are from the former group of users, this project might be for you.

## What is tc\_reader ?
The tc\_reader acts as a persistent script for the Net-SNMP agent. See man
SNMPD.CONF(5snmp) for more details on persistent scripts. It is started by the
SNMP daemon and it regularly runs *iproute2/tc* command to get all the queue and
class statistics. These statistics are then provided to the SNMP daemon whenever
SNMP GET or GET-NEXT query is made to the OID that is served by the tc\_reader.

The tc\_reader is written in Go in order to be fast and reliable.

## Installation
1.  *Get the binary:* You have two options. You can either take one of the
    pre-compiled binaries, or get the latest version of the source code directly
    from [github.com](http://github.com) and compile it yourself.

2.  *Configure tc\_reader*: Take the sample configuration file and adjust it to
    match your needs.

3.  *Configure SNMPD:* After you have the binary, you need to have a working
    Net-SNMP agent and configure the tc\_reader as a persistent script.

4.  *Add the graphs:* Lastly you might want to start graphing the statistics in
    your favorite monitoring solution. Example for [Cacti](http://www.cacti.net/)
    is provided.

### Get the binary - Method 1 - Download pre-compiled binaries
Sorry, no binary releases yet, coming soon...

### Get the binary - Method 2 - The do-it-yourself method
1.  You will need a working Go installation on your machine. On Debian/Ubuntu
    you can achieve this by running:
    ```
    apt-get install golang
    ```

2.  Next you need to set-up the *$GOPATH* environment variable. See the article
    [How to Write Go Code](http://golang.org/doc/code.html) for more details.
    What you need to do is choose a directory where all the Go code and binaries
    will be stored. Say that you choose */home/your_username/go*. Run the
    following:
    ```
    mkdir /home/$USER/go
    export GOPATH="/home/$USER/go"
    ```

3.  You can now get the latest copy of this source-code from GitHub:
    ```
    go get github.com/mum4k/tc\_reader
    ```

4.  Now to compile and install the binary run:
    ```
    go install github.com/mum4k/tc\_reader
    ```

5.  You can find the compiled binary in */home/your_username/go/bin/tc\_reader*.
    Either keep it here or move it somewhere else on your system based on your
    preference.

### Test the binary
See if it works. Your can execute the binary and write *PING* on the empty line.
Assuming that everything is OK, tc\_reader should respond *PONG*. Yeah do not
judge me, this is the SNMPD standard for communication with persistent scripts.

### Configure tc\_reader
Get the *tc\_reader.conf* file. If you downloaded one of the pre-compiled
binaries, you will find it inside the archive. If you compiled the binary
yourself, you will find it inside the *$GOPATH*. You have to copy the
tc\_reader.conf file into your */etc/* directory:
```
cp /home/$USER/go/src/github.com/mum4k/tc_reader/tc_reader.conf /etc
```

The configuration file is self-explanatory (hopefully).

### Configure SNMPD
The assumption is that you already have a working setup of Net-SNMP daemon and
the SNMP client utilities (snmpwalk, ...). The bare minimum is:
```
apt-get install snmpd snmp
```

Edit your Net-SNMP configuration file, which is usually located in
*/etc/snmp/snmpd.conf*. You have to add one line:
```
pass_persist .1.3.6.1.4.1.2021.255 /path/to/tc_reader
```

### Restart and test SNMPD
Restart your Net-SNMP daemon. On Debian/Ubuntu you could execute:
```
service snmpd restart
```

Now try to query the Net-SNMP daemon and see if everything up to here works:
```
snmpwalk -v2c -c your_community localhost .1.3.6.1.4.1.2021.255
```

### Add the graphs [Cacti](http://www.cacti.net/).
You might need to perform different steps if you are using some other software
for graphing. However if you use [Cacti](http://www.cacti.net/), you can follow
these steps to monitor individual users configured in tc\_reader.conf.

1.  Copy the snmp\_queries templates into your cacti resource directory.
    Example:
```
cp /home/$USER/go/src/github.com/mum4k/tc_reader/cacti_templates/snmp_queries/* /usr/share/cacti/resource/snmp_queries
```

2.  Login to your Cacti console, click on *Import Templates* in the menu and
    import the data query templates from this directory:
```
/home/$USER/go/src/github.com/mum4k/tc_reader/cacti_templates/data_queries
```

3.  In Cacti console, click on *Devices*, then your device and down under
    *Associated Data Queries* add a data query called *SNMP - TC Users*. Once
    this is added to your device you can click on *Verbose Query* to see if it
    works.

4.  Finally, still on the same device page click on *Create Graphs for this
    host*, on the next page under *SNMP - TC Users* select all the users you
    would like to monitor and click *Create*.

Assuming that all of this worked - enjoy the graphs.

## Support
Feel free to submit bugs or let me know if you find anything wrong or missing.
Although this is a "pet-project" so expect some delays.
