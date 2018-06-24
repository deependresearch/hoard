# HOARD Quick Start

## Install Redis and launch Redis

If not already installed, install Redis on your local machine. This is best done through the offical YUM or APT repos.

We will be using the default port for this, but you can configure it manually if you want a different port. Take a quick look at `example.json`
In our testing environments, we successfully utilized the Docker Redis image and it functioned flawlessly over a month period.


## Configure Suricata to log to Redis

Edit /etc/suricata/outputs.yaml and add the following:

  ```
  # Extensible Event Format (nicknamed EVE) event log in JSON format
  - eve-log:
      enabled: yes
      type: redis
      redis:
        enabled: yes
        server: 127.0.0.1
        port: 6379
        async: true
        mode: list
        key: suricata
      types:
        - dns
        - http
```

## Download or Build Binaries

Linux ELF Binaries (64bit) are provided in the Github Repo's DIST directory. Place them in `/opt/hoard/`; then create the directory `/opt/hoard/sketches`

Alternatively, you may compile the binaries yourself by:

* Installing Golang (https://golang.org/doc/install)

* Running: `go get github.com/seiflotfy/cuckoofilter; go get github.com/go-redis/redis`

* Then: `go build -ldflags "-s -w" hoard_server.go; go build -ldflags "-s -w" hoard_client.go`


## Enable Boot Startup and crash restart with Supervisor

Install and Configure Supervisor (Use Official YUM/APT repositories) to launch the Hoard Server application:

>/etc/supervisor/conf.d/hoard.conf
```
    [program:hoard]
    command=/opt/hoard/hoard_server
    directory=/opt/hoard
    stdout_logfile=/var/log/hoard.log
    redir_stderr=true
```

Update Supervisor configs with:  `supervisorctl update`

Verify HOARD status with: `supervisorctl status`

	hoard                            RUNNING    pid 19942, uptime 0:00:30

## Monitor for Sketches

Hoard will drop sketches in `/opt/hoard/sketches` by default. In low bandwidth environments you can expect to see a new sketch every 2 hours. Larger environments will write more frequently.

If you can't wait and want to see the benefit right away, you can try the "historical_suricata" application in DIST

## Historical Suricata Usage

Copy an eve.json file (or gzip the .gz files) to a temporary directory.
Create the directory `sketches` and run the binary `./historical_suricata -f eve.json`

You can then use the Hoard Client to quickly search for any observables and get a feel for how the application works.
If you'd like to avoid placing HOARD into Production (it is a POC after all); you could technically use the historical_suricata script to parse the files offline each day.


## Use Hoard Client to search for intelligence

Place your intelligence file in the hoard directory as `intel.txt` and simply run `./hoard_client`.

For testing purposes, you may wish to put common domains in the intel.txt file. Remember you'll need to wait for your first sketch to write before this will do anything useful.

## While you wait...

I highly recommend contributing to the Twitter debates about pineapple on pizza.
When doing so, be sure to tag `@9bplus` and `@Andrew___Morris`. I'd like to see how long it takes them to work this back. :)
