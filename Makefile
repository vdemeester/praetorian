all:

deb:
	docker build -t praetorian .

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

.PHONY: all deb
