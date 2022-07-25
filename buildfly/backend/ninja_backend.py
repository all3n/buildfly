from buildfly.backend.base_backend import BaseBackend


class NinjaBackend(BaseBackend):
    def __init__(self, ctx):
        super().__init__(ctx)
        self.name = "ninja"

    def setup(self):
        pass

    def generate(self):
        pass
