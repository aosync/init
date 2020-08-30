# init

A fast init system for Linux.

## Roles of init.

PID 1 action is separated into 3 phases.

### Phase 0

- The base filesystems (`/dev`, `/proc`, `/sys`, `/run`) are mounted.
- System entropy is seeded with a sample of last run entropy preserved at `/var/entropy`.
- Hostname is fetched from `/etc/hostname` and is set.

The rest of phase 0 is delegated to programs located at `/etc/init/0`, they are executed sequentially and in a synchronous manner to ensure everything is ready for phase 1.
A notable program of phase 0 is `mounts`, which essentially reads from fstab.

### Phase 1

Phase 1 sets up less critical and less important stuff for the userland.

Programs located at `/etc/init/1/once` are executed once and in an asynchronous manner, in the background.
Programs located at `/etc/init/1/repeat` are executed in an asynchronous manner and in the background, but they get respawned when one exits.

`1/repeat` can thus be used as a very minimal service manager, or can be used to manage another service manager such as `runsvdir`.

PID 1 is also now ready to reap orphaned child processes (zombies).

### Phase 2

Phase 2 ensures the system shutdown or reboot to be proceeded in an orderly fashion. It is triggered when PID 1 receives SIGUSR1 or SIGUSR2.

A sample of the system entropy is preserved and stored at `/var/entropy`, then programs located at `/etc/init/2` are executed in a synchronous manner. A sync(2) syscall is issued then the init calls for the shutdown or reboot.

## Installation

```sh
make
make DESTDIR="/usr" install
```

## Signals

SIGUSR1 causes shutdown and SIGUSR2 causes reboot.
