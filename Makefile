CLIB = libclib.so

all: compile-clib node-compile-gyp

compile-clib:
	pwd
	go build -v -buildmode=c-shared -o $(CLIB) relayer/clib.go
	test "`uname -s 2>/dev/null`" != Darwin || install_name_tool -id $(CLIB) $(CLIB)

node-compile-gyp:
	if yarn -v >/dev/null 2>&1; then \
		yarn build:gyp; \
	else \
		npm run build:gyp; \
	fi

# Only run from the package.json build:gyp script
compile-gyp: #create-binding-gyp
	node-gyp configure build $(GYP_DEBUG) || { status=$$?; rm -f binding.gyp; exit $$status; }
#	rm -f binding.gyp

create-binding-gyp:
	sed -e 's%@CLIBDIR@%$(CLIBDIR)%g; s%@CLIB@%$(CLIB)%g' binding.gyp.in > binding.gyp
