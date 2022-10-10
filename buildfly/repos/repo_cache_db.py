import datetime
import sqlite3
import os
import json

from buildfly.repos.common import BFlyPkg
from buildfly.utils.log_utils import get_logger
from buildfly.utils.system_utils import get_bfly_path
from buildfly.config.global_config import G_CONFIG

logger = get_logger(__name__)
repo_db_path = get_bfly_path("repo.db")


def init_repo_db():
    repo_db.execute("CREATE TABLE repos(name text, path text, desc text, ts timestamp)")
    repo_db.execute("CREATE TABLE pkgs(repo text, name text, desc text, path text, ts timestamp)")
    repo_db.execute(
        "CREATE TABLE artifacts(pkg text, ver_id text, build_params text, hash text, os text, arch text, ts timestamp)")


def dict_factory(cursor, row):
    col_names = [col[0] for col in cursor.description]
    return {key: value for key, value in zip(col_names, row)}


if not os.path.exists(repo_db_path):
    repo_db = sqlite3.connect(repo_db_path)
    init_repo_db()
else:
    repo_db = sqlite3.connect(repo_db_path)


class RepoCacheDb(object):
    def __init__(self, db):
        self.db = db
        self.db.row_factory = dict_factory
        self.c = self.db.cursor()

    def __del__(self):
        self.c.close()

    def get_pkg(self, name, repo=None):
        pkgs = self.c.execute("""
        SELECT r.path, r.name rname, p.name pname, r.path rpath, p.path path,p.ts ts FROM `pkgs` p INNER JOIN `repos` r ON p.repo = r.name  WHERE p.`name` = ?
        """, (name,))
        pres = pkgs.fetchall()
        if pres is None:
            return None
        else:
            mf = None
            res = pres[0]
            if len(pres) > 1:
                logger.info("%d found,choose %s", len(pres), res)
            mf_json = os.path.join(res["rpath"], res['path'], 'manifest.json')
            if os.path.exists(mf_json):
                with open(mf_json, "r") as f:
                    mf = json.loads(f.read())
            return mf

    def add_pkg(self, bpkg: BFlyPkg):
        params_str = json.dumps(bpkg.params)
        self.db.execute(f"INSERT INTO `artifacts`(pkg, ver_id, build_params, hash, os, arch, ts) VALUES(?, ?, ?, ?, ?)",
                        (bpkg.name, bpkg.version, params_str, bpkg.commit_sha, bpkg.os, bpkg.arch,
                         datetime.datetime.now()))
        self.db.flush()

    def update_repo(self, repo, repo_dir, dirs):
        # c = self.db.cursor()
        rres = self.c.execute("SELECT * FROM `repos` WHERE `name`=?", (repo,))
        row = rres.fetchone()
        if row is None:
            self.db.execute(f"INSERT INTO `repos` VALUES(?, ?, ?, ?)",
                            (repo, repo_dir, repo, datetime.datetime.now()))
        else:
            # name, path, desc, ts = row
            path = row['path']
            if repo_dir != path:
                self.db.execute(f"UPDATE `repos` SET path=? WHERE name=?",
                                (repo_dir, repo))

        for d in dirs:
            manifest_file = os.path.join(repo_dir, d, "manifest.json")
            with open(manifest_file, "r") as f:
                meta = json.loads(f.read())
                name = meta["name"]
                desc = meta.get("description", "")
                path = d
                res = self.db.execute(f"SELECT * FROM `pkgs` WHERE name=?", (name,))
                pkg = res.fetchone()
                if pkg is None:
                    self.c.execute(f"INSERT INTO `pkgs`(repo, name, desc, path, ts) VALUES(?, ?, ?, ?, ?)",
                                   (repo, name, desc, path, datetime.datetime.now()))
                else:
                    print(pkg)
        # c.close()
        self.db.commit()


repo_cache = RepoCacheDb(repo_db)
