# Primarily used for runing GUI in WSL2 on Windows 10 (make sure X Server is started. WSL2 does not support GUI out of the box)
export DISPLAY_NUMBER="0.0"
export DISPLAY_IP=$(tail -1 /etc/resolv.conf | cut -d' ' -f2)
export DISPLAY=$DISPLAY_IP:$DISPLAY_NUMBER
# LIBGL_ALWAYS_INDIRECT should be 1 by default except for NVidia when it should be 0
export LIBGL_ALWAYS_INDIRECT=0
export XDG_RUNTIME_DIR=$PATH:~/.cache/xdgr
# export PULSE_SERVER=${PULSE_SERVER:-tcp:$DISPLAY_IP}  # SOUND SERVER


# if running under MSYS2 or MingW then execute the exe
if [[ -n $MSYSTEM ]]; then
    # run app
    exec ./player.exe -f=./demo.mp4 -a=false -fw=true; # WSL2 doesn't have sound driver therefore set audio flag -a to false
else
    # if executed in WSL, then make sure X Server is started 
    if [[ -n "$IS_WSL" || -n "$WSL_DISTRO_NAME" ]]; then
        if ! timeout 1s xset q &>/dev/null; then
            echo "No X server at \$DISPLAY [$DISPLAY]" >&2
            exit 1    
        fi
    fi
    # run app
    exec ./player -f=./demo.mp4 -a=false -fw=true; # WSL2 doesn't have sound driver therefore set audio flag -a to false
fi
