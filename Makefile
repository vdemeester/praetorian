VERSION=0.1.0
BUILD=1

prefix=/usr/local
bindir=${prefix}/bin
mandir=${prefix}/share/man

all:

clean:
	rm -rf *.deb debian man/man*/*.html
	find . -name '*~' -delete

install: install-bin install-man

install-bin:
	install -d $(DESTDIR)$(bindir)
	find bin -type f -printf %P\\0 | xargs -0r -I__ install bin/__ $(DESTDIR)$(bindir)/__

install-man:
	find man -type d -printf %P\\0 | xargs -0r -I__ install -d $(DESTDIR)$(mandir)/__
	find man -type f -name \*.[12345678] -printf %P\\0 | xargs -0r -I__ install -m644 man/__ $(DESTDIR)$(mandir)/__
	find man -type f -name \*.[12345678] -printf %P\\0 | xargs -0r -I__ gzip $(DESTDIR)$(mandir)/__

uninstall: uninstall-bin uninstall-man

uninstall-bin:
	find bin -type f -printf %P\\0 | xargs -0r -I__ rm -f $(DESTDIR)$(bindir)/__
	rmdir -p --ignore-fail-on-non-empty $(DESTDIR)$(bindir) || true

uninstall-man:
	find man -type f -name \*.[12345678] -printf %P\\0 | xargs -0r -I__ rm -f $(DESTDIR)$(mandir)/__.gz
	find man -depth -mindepth 1 -type d -printf %P\\0 | xargs -0r -I__ rmdir $(DESTDIR)$(mandir)/__ || true
	rmdir -p --ignore-fail-on-non-empty $(DESTDIR)$(mandir) || true

build:
	make install prefix=/usr sysconfdir=/etc DESTDIR=debian
	fpm -s dir -t deb \
		-n praetorian -v $(VERSION) --iteration $(BUILD) -a all \
		-d coreutils -d dash -d dpkg -d openssh-server -d util-linux \
		-m "Vincent Demeester <vincent@sbr.pm>" \
		--url "https://github.com/vdemeester/praetorian" \
		--description "A ssh praetorian (bouncer, minder or whatever) ; it's just a cool restricted command script.." \
		-C debian .
	make uninstall prefix=/usr sysconfdir=/etc DESTDIR=debian

man:
	find man -name \*.ronn | xargs -n1 ronn --manual=Freight --style=toc

#docs:
#	for SH in $$(find bin lib -type f -not -name \*.html); do \
#		shocco $$SH >$$SH.html; \
#	done

gh-pages: 
	mkdir -p gh-pages
	curl -X POST --data name=praetorian --data theme=v1 --data-urlencode content@README.md http://documentup.com/compiled > gh-pages/index.html
	git checkout -q origin/gh-pages -B gh-pages
	cp -R gh-pages/* ./
	rm -rf gh-pages
	git add .
	git commit -m "Rebuilt manual."
	git push origin gh-pages
	git checkout -q master

.PHONY: all install uninstall build # man gh-pages
