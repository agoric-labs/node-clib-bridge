CLIBDIR = <!(pwd)
CLIB = libclib.so

all: compile-go node-compile-gyp

compile-go:
	go build -v -buildmode=c-shared -o libclib.so clib.go
	test "`uname -s 2>/dev/null`" != Darwin || install_name_tool -id `pwd`/libclib.so libclib.so

node-compile-gyp:
	if yarn -v >/dev/null 2>&1; then \
		yarn build:gyp; \
	else \
		npm run build:gyp; \
	fi

# Only run from the package.json build:gyp script
compile-gyp:
	sed -e 's%@CLIBDIR@%$(CLIBDIR)%g; s%@CLIB@%$(CLIB)%g' binding.gyp.in > binding.gyp
	node-gyp configure build $(GYP_DEBUG) || { status=$$?; rm -f binding.gyp; exit $$status; }
	rm -f binding.gyp
