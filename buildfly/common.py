import glob


class BFlyRepo(object):
    pass


class BFlyGithubRepo(BFlyRepo):
    pass


class BFlyHttpRepo(BFlyRepo):
    pass


class BFlyLocalRepo(BFlyRepo):
    pass


import json


class BFlyManifest(object):
    def __init__(self, name=None, mode=None, repo={}, version=None, description=None, envs={}, configs={},
                 vars={}, arch=None, system=None, prefix=None, libc_version=None):
        self.name = name
        self.mode = mode
        self.repo = repo  # name, url, branch, commit, tag
        self.version = version
        self.description = description
        self.envs = envs
        self.configs = configs
        self.vars = vars
        self.arch = arch
        self.system = system
        self.prefix = prefix
        self.libc_version = libc_version

    def to_json(self, indent=4):
        return json.dumps(self.__dict__, indent=indent)

    @staticmethod
    def from_dict(v):
        bm = BFlyManifest()
        bm.__dict__ = v
        return bm

    @staticmethod
    def from_json(v):
        bm = BFlyManifest()
        bm.__dict__ = json.loads(v)
        return bm


class BFlyDep(object):
    def __init__(self, name, artifact_id=None, repo="github", url=None, version=None, cmds=[], modules=[], sha256=None):
        """
        repo:
            github
            url
        """
        self.name = name
        self.artifact_id = artifact_id
        self.repo = repo
        self.url = url
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

    def get_all_srcs(self):
        srcs = []
        if type(self.srcs) == str:
            srcs.extend(glob.glob(self.srcs, recursive=True))
        else:
            for src in self.srcs:
                srcs.extend(glob.glob(src, recursive=True))
        return srcs

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
