from buildfly.actions.base_action import BaseAction


class NewAction(BaseAction):
    def parse_args(self, parser):
        pass

    def run(self):
        super().run()
