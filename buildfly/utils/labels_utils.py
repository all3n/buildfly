'''
//my/app/main:app_binary

// is root prefix
my/app/main: package name
app_binary: target_name

: is ommitted   target name is equal last part of package name
//my/app
//my/app:app

relative define:
//my/app:BUILDFLY

//my/app:app
//my/app
:app
app



'''

def parse_label(label: str, package):
    if label.startswith("//"):
        # is abs path
        if ":" in label:
            # full path with target name
            return label
        else:
            # target name is ommited, set target equal last part of package name
            return "%s/%s" % (label, label.split("/")[-1])
    else:
        # relative label
        assert package.startswith("//"), "package must abs path"
        target_name = label.split(":")[-1]
        return "%s:%s" % (package, target_name)
