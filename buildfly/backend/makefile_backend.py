from buildfly.backend.base_backend import BaseBackend


class MakefileBackend(BaseBackend):
    def __init__(self, ctx):
        super().__init__(ctx)
        self.name = "makefile"

    def setup(self):
        pass

    def generate(self):
        pass
