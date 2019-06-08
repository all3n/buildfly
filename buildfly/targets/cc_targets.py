from buildfly.build.build_manager import BUILD_MANAGER


class CcTarget(object):
    def __init__(self, name,
                 target_type,
                 deps = None,
                 srcs= None,
                 hdrs = None,
                 copts = None,
                 linkopts = None):
        self.name = name
        self.target_type = target_type
        self.deps = deps
        self.srcs = srcs
        self.hdrs = hdrs
        self.copts = copts
        self.linkopts = linkopts

        assert len(srcs) > 0, "srcs must be set"

    def __str__(self):
        return "{cc_target[%s] name:%s deps:%s}" % (self.target_type, self.name, self.deps)


def cc_binary(
        name,
        deps=None,
        srcs=None,
        copts=None,
        hdrs=None,
        linkopts=None,
        linkshared=False,
        linkstatic=True
):
    target = CcTarget(
        target_type="bin",
        name=name,
        deps=deps,
        srcs=srcs,
        hdrs=hdrs,
        copts = copts,
        linkopts = linkopts
    )
    BUILD_MANAGER.register_target(target)


def cc_library(
        name,
        deps=None,
        srcs=None,
        hdrs=None,
        copts=None,
        linkopts=None
):
    target = CcTarget(
        target_type="lib",
        name=name,
        deps=deps,
        srcs=srcs,
        hdrs=hdrs,
        copts = copts,
        linkopts = linkopts
    )
    BUILD_MANAGER.register_target(target)


def cc_import(name,
              hdrs=None,
              static_library=None,
              shared_library=None,
              alwayslink=False
              ):
    pass

def cc_test(name):
    pass


BUILD_MANAGER.register_func(cc_binary)
BUILD_MANAGER.register_func(cc_library)
BUILD_MANAGER.register_func(cc_import)
BUILD_MANAGER.register_func(cc_test)
