{
  "targets": [
    {
      "target_name": "clib",
      "cflags!": [ "-fno-exceptions" ],
      "cflags_cc!": [ "-fno-exceptions" ],
      "xcode_settings": {
        'GCC_ENABLE_CPP_EXCEPTIONS': 'YES',
        'CLANG_CXX_LIBRARY': 'libc++',
        'MACOSX_DEPLOYMENT_TARGET': '10.7',
      },
      "msvs_settings": {
        "VCCLCompilerTool": { "ExceptionHandling": 1 },
      },
      "sources": [ "bridge-node.cc", "<!(pwd)/libclib.so" ],
      "include_dirs": [
        "<!@(node -p \"require('node-addon-api').include\")",
        "<!@(node -p \"require('napi-thread-safe-callback').include\")"
      ],
      "dependencies": ["<!(node -p \"require('node-addon-api').gyp\")"],
      "libraries": [ "-Wl,-rpath,<!(pwd)", "<!(pwd)/libclib.so" ],
      "conditions": [
        ['OS=="mac"', {
          'cflags+': ['-fvisibility=hidden'],
          'xcode_settings': {
            'GCC_SYMBOLS_PRIVATE_EXTERN': 'YES', # -fvisibility=hidden
          }
        }],
        ['OS=="win"', {
          'defines': [ '_HAS_EXCEPTIONS=1' ]
        }]
      ],
      "actions": [
        {
            'action_name': 'compile-clib',
            'inputs': ['Makefile', 'relayer/clib.go'],
            'outputs': ['<!(pwd)/libclib.so'],
            'action': ['make', 'compile-clib', 'CLIB=<!(pwd)/libclib.so'],
        }
      ],
    },
  ],
}
