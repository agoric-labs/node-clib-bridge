#include <napi.h>
#include <napi-thread-safe-callback.hpp>
#include <iostream>

#include <stdlib.h>
typedef const char* Body;
typedef int (*sendFunc)(int, int, Body);

typedef long long GoInt64;
typedef GoInt64 GoInt;
typedef struct { void *data; GoInt len; GoInt cap; } GoSlice;

extern "C" {
  extern int RunClib(int nodePort, sendFunc p1, GoSlice args);
  extern int ReplyToClib(int replyPort, int isRejection, Body p2);
  extern Body SendToClib(int instance, Body p1);
}

namespace NodeBridge {

static std::shared_ptr<ThreadSafeCallback> dispatcher;

class NodeReply {
public:
    NodeReply(bool isRejection, std::string value) :
        _isRejection(isRejection), _value(value) {}
    std::string value() {
        return _value;
    }
    bool isRejection() {
        return _isRejection;
    }
private:
    bool _isRejection;
    std::string _value;
};

class NodeReplier : public Napi::ObjectWrap<NodeReplier> {
public:
    static Napi::Object Init(Napi::Env env, Napi::Object exports) {
        Napi::HandleScope scope(env);
        Napi::Function func = DefineClass(env, "NodeReplier", {
            InstanceMethod("resolve", &NodeReplier::Resolve),
            InstanceMethod("reject", &NodeReplier::Reject),
        });

        constructor = Napi::Persistent(func);
        constructor.SuppressDestruct();

        exports.Set("NodeReplier", func);
        return exports;
    }

    NodeReplier(const Napi::CallbackInfo& info) : Napi::ObjectWrap<NodeReplier>(info) {
        // Do nothing, we're initialized by New.
    }

    static Napi::Object New(Napi::Env env, std::shared_ptr<std::promise<NodeReply>> promise) {
        Napi::Object obj = constructor.New({});
        NodeReplier* self = Unwrap(obj);
        self->_promise.swap(promise);
        return obj;
    }

private:
    static Napi::FunctionReference constructor;
    void doReply(bool isRejection, std::string value) {
        NodeReply reply(isRejection, value);
        _promise->set_value(reply);
    }

    void Resolve(const Napi::CallbackInfo& info) {
        doReply(false, info[0].As<Napi::String>().Utf8Value());
    }

    void Reject(const Napi::CallbackInfo& info) {
        doReply(true, info[0].As<Napi::String>().Utf8Value());
    }

    std::shared_ptr<std::promise<NodeReply>> _promise;
};

Napi::FunctionReference NodeReplier::constructor;

int SendToNode(int port, int replyPort, Body str) {
    //std::cerr << "Send to node port " << port << " " << str << std::endl;
    std::string instr(str);
    std::thread([instr, port, replyPort]{
        auto promise = std::make_shared<std::promise<NodeReply>>();
        dispatcher->call(
            // Prepare arguments.
            [port, instr, promise](Napi::Env env, std::vector<napi_value>& args){
                // std::cerr << "Calling threadsafe callback with " << instr << std::endl;
                args = {
                    Napi::Number::New(env, port),
                    Napi::String::New(env, instr),
                    NodeReplier::New(env, promise),
                };
            });
        // std::cerr << "Waiting on future" << std::endl;
        try {
            NodeReply ret = promise->get_future().get();
            // std::cerr << "Replying to clib with " << ret.value() << " " << ret.isRejection() << std::endl;
            if (replyPort) {
              ReplyToClib(replyPort, ret.isRejection(), ret.value().c_str());
            }
        } catch (std::exception& e) {
            // std::cerr << "Exceptioning " << e.what() << std::endl;
            if (replyPort) {
              ReplyToClib(replyPort, true, e.what());
            }
        }
        // std::cerr << "Thread is finished" << std::endl;
    }).detach();
    // std::cerr << "Ending Send to Node " << str << std::endl;
    return 0;
}

static Napi::Value sendClib(const Napi::CallbackInfo& info) {
    // std::cerr << "Send to clib" << std::endl;
    Napi::Env env = info.Env();
    int instance = info[0].As<Napi::Number>();
    std::string tmp = info[1].As<Napi::String>().Utf8Value();
    Body ret = SendToClib(instance, tmp.c_str());
    return Napi::String::New(env, ret);
}

static Napi::Value runClib(const Napi::CallbackInfo& info) {
    Napi::Env env = info.Env();
    // std::cerr << "Starting clib from Node Controller" << std::endl;

    int nodePort = info[0].As<Napi::Number>().ToNumber();
    dispatcher = std::make_shared<ThreadSafeCallback>(info[1].As<Napi::Function>());
    Napi::Array clibArgv = info[2].As<Napi::Array>();
    unsigned int argc = clibArgv.Length();
    char** argv = new char*[argc];
    for (unsigned int i = 0; i < argc; i ++) {
        if (clibArgv.Has(i)) {
            std::string tmp = clibArgv.Get(i).As<Napi::String>().Utf8Value();
            argv[i] = strdup(tmp.c_str());
        } else {
            argv[i] = nullptr;
        }
    }

    GoSlice args = {argv, argc, argc};
    int clibPort = RunClib(nodePort, SendToNode, args);

    for (unsigned int i = 0; i < argc; i ++) {
        free(argv[i]);
    }
    delete[] argv;
    // std::cerr << "End of starting clib from Node " << clibPort << std::endl;
    return Napi::Number::New(env, clibPort);
}

static Napi::Object InitAll(Napi::Env env, Napi::Object exports) {
    exports = NodeReplier::Init(env, exports);
    exports.Set(
        Napi::String::New(env, "runClib"),
        Napi::Function::New(env, runClib, "runClib"));
    exports.Set(
        Napi::String::New(env, "sendClib"),
        Napi::Function::New(env, sendClib, "sendClib"));
    return exports;
}

NODE_API_MODULE(NODE_GYP_MODULE_NAME, InitAll)

}
