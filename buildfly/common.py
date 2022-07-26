import glob
from buildfly.utils.yaml_conf_utils import BuildDependency
from buildfly.utils.dep_utils import get_dep

class BFlyRepo(object):
    pass


class BFlyGithubRepo(BFlyRepo):
    pass


class BFlyHttpRepo(BFlyRepo):
    pass


class BFlyLocalRepo(BFlyRepo):
    pass


class BFlyDep(object):
    def __init__(self, name, artifact_id = None, repo="github", url = None, version=None, cmds=[], modules=[], sha256=None):
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

    def install_if_needed(self):
        if self.repo == "github":
            dep_obj = None
            if self.repo == "github" and self.artifact_id:
                dep_obj = f"{self.artifact_id}@{self.version}" if self.version else f"{self.artifact_id}"
            elif self.url:
                dep_obj = {}
                if self.modules:
                    dep_obj['modules'] = self.modules
                    dep_obj["cmds"] = self.cmds
                    dep_obj["url"] = self.url
            bdep = BuildDependency(self.name, dep_obj=dep_obj)
            get_dep(bdep)



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
