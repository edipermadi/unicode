WGET	:= $(shell which wget)

.PHONY: data
data: data/ucd.all.flat.zip data/ucd.all.grouped.zip

data/ucd.all.flat.zip:
	$(WGET) -c https://www.unicode.org/Public/UCD/latest/ucdxml/ucd.all.flat.zip -O $@

data/ucd.all.grouped.zip:
	$(WGET) -c https://www.unicode.org/Public/UCD/latest/ucdxml/ucd.all.grouped.zip -O $@ 