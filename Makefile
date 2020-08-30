init:
	go build .

install: init
	cp -R etc/init ${PREFIX}/etc
	cp -f bin/* ${DESTDIR}/bin
	cp init ${DESTDIR}/bin/inao
	ln -sf inao ${DESTDIR}/bin/init
