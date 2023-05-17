#!/bin/sh
#
# Startup script for Main
export APPNAME="futures-trading"
export CONF=config.dev.toml
export ENV=.env
pid_file=tmp/$APPNAME.pid
log_file=tmp/$APPNAME.log

# Arguments to pass to the service
PARAMS=" $1 \
        --config $CONF --env $ENV"

echo $PARAMS

is_running () {
    [ -f "$pid_file" ] && ps `cat $pid_file` > /dev/null 2>&1
}

case "$1" in
    start)
        # Main startup
        echo -n "Starting $APPNAME: "
        exec ./$APPNAME $PARAMS > $log_file 2>&1 &
        [ ! -z $pid_file ] && echo $! > $pid_file
        echo "OK. Check your stdout logs"
        ;;
    stop)
        # Main shutdown
        if ! is_running; then
            echo "Service stopped"
            exit 1
        fi
        echo -n "Shutdown $APPNAME: "
        while is_running;
        do
            kill `cat $pid_file`
            sleep 1
        done
        echo "OK"
        ;;
    reload|restart)
        $0 stop
        $0 start 
        ;;
    status)
        if is_running; then
            echo -n "Service is running. Pid: "
            echo `cat $pid_file`
        else
            echo "Service stopped"
        fi
        ;;
    *)
        echo "Usage: `basename $0` start|stop|restart|reload"
        exit 1
esac

exit 0