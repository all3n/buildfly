class BFlyRepo(object):
    pass


class BFlyGithubRepo(BFlyRepo):
    pass


class BFlyHttpRepo(BFlyRepo):
    pass


class BFlyLocalRepo(BFlyRepo):
    pass


class BFlyDep(object):
    def __init__(self, name, repo="github", version=None, cmds=[], modules=[], sha256=None):
        """
        repo:
            github
            url
        """
        self.name = name
        self.repo = repo
        self.version = version
        self.cmds = []
        self.modules = modules


class BFlyTarget(object):
    def __init__(self, **kwargs):
        self.name = ""
        self.srcs = []
        self.cflags = []
        self.cxxflags = []
        self.ldflags = []
        self.includes = []
        self.library = []
        self.library_dirs = []
        self.rpath = []
        self.static_gcc = False
        self.mode = "Debug"  # Release
        self.defines = []
        self.deps = []

    def init(self, **kwargs):
        for k, v in kwargs.items():
            setattr(self, k, v)

    def __repr__(self):
        return "[%s][%s]{%s}" % (self.__class__.__name__, self.name, str(self.__dict__))


class BFlyBin(BFlyTarget):

    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        self.init(**kwargs)


class BFlyLibrary(BFlyTarget):

    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        self.link_type = "shared"
        self.init(**kwargs)
