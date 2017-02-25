#!/usr/bin/env bin/python

import os, os.path, yaml, json, string, subprocess, traceback, pprint
from dateutil import parser


"""
# find final url
def follow_redirect(url):
    print "following", url, "..."
    httplib.HTTPConnection.debuglevel = 1
    request = urllib2.Request(url)
    opener = urllib2.build_opener()
    f = opener.open(request)
    return f.url
"""

# --------------------------------------------- UTILITY BELT -----------------------------------------------------------

def walklevel(some_dir, level=1):
    some_dir = some_dir.rstrip(os.path.sep)
    assert os.path.isdir(some_dir)
    num_sep = some_dir.count(os.path.sep)
    for root, dirs, files in os.walk(some_dir):
        yield root, dirs, files
        num_sep_this = root.count(os.path.sep)
        if num_sep + level <= num_sep_this:
            del dirs[:]


# http://stackoverflow.com/questions/956867/how-to-get-string-objects-instead-of-unicode-ones-from-json-in-python/6633651#6633651
# http://stackoverflow.com/questions/9590382/forcing-python-json-module-to-work-with-ascii
def _decode_list(data):
    rv = []
    for item in data:
        if isinstance(item, unicode):
            item = item.encode('utf-8')
        elif isinstance(item, list):
            item = _decode_list(item)
        elif isinstance(item, dict):
            item = _decode_dict(item)
        rv.append(item)
    return rv


def _decode_dict(data):
    rv = {}
    for key, value in data.iteritems():
        if isinstance(key, unicode):
            key = key.encode('utf-8')
        if isinstance(value, unicode):
            value = value.encode('utf-8')
        elif isinstance(value, list):
            value = _decode_list(value)
        elif isinstance(value, dict):
            value = _decode_dict(value)
        rv[key] = value
    return rv


# for "object_pairs_hook=deunicodify_hook"
# http://stackoverflow.com/questions/956867/how-to-get-string-objects-instead-of-unicode-ones-from-json-in-python/34796078#34796078
def deunicodify_hook(pairs):
    new_pairs = []
    for key, value in pairs:
        if isinstance(value, unicode):
            value = value.encode('utf-8')
        if isinstance(key, unicode):
            key = key.encode('utf-8')
        new_pairs.append((key, value))
    return dict(new_pairs)


# http://www.bogotobogo.com/python/python_longest_common_substring_lcs_algorithm_generalized_suffix_tree.php
def lcs(S, T):
    m = len(S)
    n = len(T)
    counter = [[0]*(n+1) for x in range(m+1)]
    longest = 0
    lcs_set = set()
    for i in range(m):
        for j in range(n):
            if S[i] == T[j]:
                c = counter[i][j] + 1
                counter[i+1][j+1] = c
                if c > longest:
                    lcs_set = set()
                    longest = c
                    lcs_set.add(S[i-c+1:i+1])
                elif c == longest:
                    lcs_set.add(S[i-c+1:i+1])
    return "".join(lcs_set)


def check_git_exist(rootpath="", pkg=""):
    return os.path.exists(os.path.join(os.path.join(rootpath, pkg), ".git"))


# --------------------------------------------- DEPENDENCY PARSER ------------------------------------------------------


# parsing Godeps.json
def parse_godeps(dependencies=None):
    rpkg = dict()
    subpkg = list()

    for dep in dependencies:
        rev = string.strip(dep["Rev"]).encode('ascii', 'replace')
        pkg = string.strip(dep["ImportPath"]).encode('ascii', 'replace')
        if rev not in rpkg:
            rpkg[rev] = list()
        rpkg[rev].append(pkg)

    for (ver, ips) in rpkg.iteritems():
        # find the repository
        repo_name = ips[0]
        if 1 < len(ips):
            for ip in ips:
                repo_name = lcs(repo_name, ip)
        if repo_name.endswith("/"):
            repo_name = repo_name[:-1]
        # in case repo is pointing subdir as in Godeps, reduce it to point to repo itself
        if 3 < len(repo_name.split("/")):
            repo_name = "/".join(repo_name.split("/")[:3])

        # print result
        print ver
        if 1 < len(ips):
            for ip in ips:
                print "\t", ip
        print "\t", "-" * 16
        print "\t", repo_name, "\n"

        subpkg.append((repo_name, ver))

    return subpkg


# parsing glide.yaml
def parse_glide(dependencies=None):
    rpkg = dict()
    subpkg = list()

    for dep in dependencies:
        rev = string.strip(dep["version"]).encode('ascii', 'replace')
        pkg = string.strip(dep["package"]).encode('ascii', 'replace')

        if pkg not in subpkg:
            subpkg.append((pkg, rev))
        if rev not in rpkg:
            rpkg[rev] = pkg

    for (ver, pkg) in rpkg.iteritems():
        print "\n", ver
        print "\t", "-" * 16
        print "\t", pkg

    return subpkg


# parsing vendor.sh
def parse_vendoer_sh(dependencies=None):
    rpkg = dict()
    subpkg = list()

    for dep in dependencies:
        if dep.startswith("clone git"):
            (pkg, rev) = dep.split(" ")[2:4]
            pkg = string.strip(pkg).encode('ascii', 'replace')
            rev = string.strip(rev).encode('ascii', 'replace')
            if pkg not in subpkg:
                subpkg.append((pkg, rev))
            if rev not in rpkg:
                rpkg[rev] = pkg

    for (ver, pkg) in rpkg.iteritems():
        print "\n", ver
        print "\t", "-" * 16
        print "\t", pkg

    return subpkg


# --------------------------------------------- SORTING, FILTERING -----------------------------------------------------


# Sort entire packages
def sort_packages(packages=None, subpkgs=None, origin=None, coalesce=False):

    for (pkg, ver) in subpkgs:
        if pkg not in packages:
            packages[pkg] = list()

        # Even if two component share the same version, we'd like to expose the shared ones and put them in GOPATH.
        # if coalesce == false, then we'd like to separate them
        if coalesce:
            packages[pkg].append((ver, origin))
        else:
            has_found = False
            for (v, o) in packages[pkg]:
                if v == ver:
                    has_found = True
                    break
            if not has_found:
                packages[pkg].append((ver, origin))


def nonconflict_mirror_command(rootpath=None, pkg=None, ver=None):
    repodir = os.path.join(rootpath, pkg)
    if pkg.startswith("github.com/"):
        stubdir = os.path.join(rootpath, "/".join(pkg.split("/")[:-1]))
        return "mkdir -p {} && cd {} && git clone https://{} && cd {} && git checkout {}".format(stubdir, stubdir, pkg,
                                                                                                 repodir, ver)
    else:
        return "(go get -d {} || true) && cd {} && git checkout {}".format(pkg, repodir, ver)


def conflict_mirror_command(rootpath=None, pkg=None):
    repodir = os.path.join(rootpath, pkg)
    if pkg.startswith("github.com/"):
        stubdir = os.path.join(rootpath, "/".join(pkg.split("/")[:-1]))
        return "/bin/mkdir -p {} && cd {} && /usr/local/bin/git clone --branch master https://{} && cd {}".format(stubdir, stubdir, pkg, repodir)
    else:
        return "(/usr/local/Cellar/go/1.7.5/libexec/bin/go get -d {} >/dev/null 2>&1 || true) && cd {}".format(pkg, repodir)


def checkout_commit(rootpath=None, pkg=None, ver=None):
    repodir = os.path.join(rootpath, pkg)
    print "{} chechking out to commit {}".format(pkg, ver)
    subprocess.call("cd {} && git checkout {}".format(repodir, ver), shell=True)
    print "\n"


def find_latest_commit(rootpath=None, pkg=None, versions=None):
    fullpath = os.path.join(rootpath, pkg)
    if not os.path.exists(fullpath + '/.git'):
        print "{} !!!NO GIT FOUND!!!! (This must be a main component package)\n".format(pkg)
        return

    vertime = list()
    for (ver, origin) in versions:
        # we can also use "cd {} && /usr/local/bin/git show -s --format=%ci {}", but this generates error on tag
        # "git log -1 --format=%ai {}" error sometimes
        # 'git log -1 --simplify-by-decoration --pretty="format:%ci"'
        p = subprocess.Popen("cd {} && /usr/local/bin/git log -1 --format=%ai {}".format(fullpath, ver), stdin=subprocess.PIPE,
                             stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True)
        output, err = p.communicate(b"input data that is passed to subprocess' stdin")
        try:
            commit_time = parser.parse(output)
        except Exception as e:
            print str(e)
            traceback.print_exc()
            print "TIME ERROR!!!! [", output, "]"
            commit_time = parser.parse("2000-01-01 00:00:00 +0000")
        finally:
            vertime.append((ver, origin, commit_time))

    commit_sorted = sorted(vertime, key=lambda t: t[2], reverse=True)
    print "{} chechking out to commit {}".format(pkg, commit_sorted[0][0])
    subprocess.call("cd {} && /usr/local/bin/git checkout {}".format(fullpath, commit_sorted[0][0]), shell=True)
    print "\n"
    return commit_sorted


# ----------------------------------------------------- MAIN -----------------------------------------------------------


if __name__ == "__main__":

    WORK_ROOT = os.environ["WORK_ROOT"]
    GOREPO = os.environ["GOREPO"]

    pkg_root = "{}/src".format(GOREPO)
    # these are the main component we need to manually keep
    main_compoment = ["github.com/coreos/etcd",
                      "github.com/docker/libcompose",
                      "github.com/docker/swarm",
                      "github.com/docker/distribution",
                      "github.com/gravitational/teleport"]

    # these are the component that needs packages
    single_vendor_required = ["etcd", "pocketcluster"]

    # package basket
    package = dict()

    print "-" * 16, "Condensing 3rd party dependencies...", "-" * 16
    print "\n\n Environments : GOREPO {}, WORK_ROOT {}".format(GOREPO, WORK_ROOT)

    for (path, dirs, files) in walklevel(os.path.join(WORK_ROOT, "asset"), 0):
        for filename in files:
            fullpath = os.path.join(path, filename)
            print "\n", "-" * 8, fullpath, "-" * 8
            with open(fullpath, "r") as depfile:
                # Godep
                if fullpath.endswith("Godeps.json"):
                    origin = filename.replace("-Godeps.json", "")
                    godeps = parse_godeps(json.load(depfile, object_hook=_decode_dict)["Deps"])
                    sort_packages(package, godeps, origin, True)
                # Glide
                elif fullpath.endswith("glide.yaml"):
                    origin = filename.replace("-glide.yaml", "")
                    glide = parse_glide(yaml.load(depfile)["import"])
                    sort_packages(package, glide, origin, True)
                # Vendor.sh
                elif fullpath.endswith("vendor.sh"):
                    origin = filename.replace("-vendor.sh", "")
                    vender_sh = parse_vendoer_sh(depfile)
                    sort_packages(package, vender_sh, origin, True)


    # print json.dumps(package)
    print "-" * 8, "downloading packages, finding latest commit", "-" * 8
    with open(os.path.join(WORK_ROOT, "dependencies.list"), "w") as finaldep:

        for pkg, vers in package.iteritems():
            repo_url = "https://{}".format(pkg)
            if 1 == len(vers):
                (version, origin) = vers[0]
                mirror_cmd = nonconflict_mirror_command(pkg_root, pkg, version)

                # check if this package is required to be installed by component
                for comp in single_vendor_required:
                    if origin.startswith(comp):
                        if not check_git_exist(pkg_root, pkg):
                            subprocess.call(mirror_cmd, shell=True)
                        else:
                            checkout_commit(pkg_root, pkg, version)

                finaldep.write("\n{}\n".format(pkg))
                finaldep.write("\t{0: <40}\t -> \t{1}\n".format(version, origin))
                finaldep.write("\t-  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -\n")
                finaldep.write("\t{}\n".format(mirror_cmd))

        for pkg, vers in package.iteritems():
            if 1 < len(vers):

                commit_sorted = None
                if pkg not in main_compoment:
                    mirror_cmd = conflict_mirror_command(pkg_root, pkg)
                    if not check_git_exist(pkg_root, pkg):
                        subprocess.call(mirror_cmd, shell=True)
                    commit_sorted = find_latest_commit(pkg_root, pkg, vers)

                finaldep.write("\n{}\n".format(pkg))
                if pkg in main_compoment:
                    finaldep.write("----------- [[ MAIN PACKGE CONFLICT ]] --------------- \n")
                else:
                    finaldep.write("----------------- !!!CONFLICT!!! --------------------- \n")

                if commit_sorted:
                    for v in commit_sorted:
                        (version, origin, comm_date) = v
                        finaldep.write("\t{0: <40}\t -> \t{1} : {2}\n".format(version, comm_date, origin))
                else:
                    for v in vers:
                        (version, origin) = v
                        finaldep.write("\t{0: <40}\t -> \t{1}\n".format(version, origin))

                if pkg not in main_compoment:
                    finaldep.write("\t-  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -\n")
                    finaldep.write("\t{}\n".format(mirror_cmd))

        finaldep.write("\n---------------- DELETE CONFLICETED VENDORS ---------------- \n")
        for pkg, vers in package.iteritems():
            if 1 < len(vers):
                finaldep.write("rm -rf {}\n".format(pkg))

        for pkg in main_compoment:
            finaldep.write("rm -rf {}\n".format(pkg))

        print "-" * 8, "cleanup vendor script", "-" * 8
        with open(os.path.join(WORK_ROOT, "vendor_cleanup.sh"), "w") as cleandep:
            cleandep.write("#!/usr/bin/env bash\n\nfunction clean_vendor() {\n")
            for pkg, vers in package.iteritems():
                if 1 < len(vers):
                    cleandep.write("\trm -rf {}\n".format(pkg))

            for pkg in main_compoment:
                cleandep.write("\trm -rf {}\n".format(pkg))

            cleandep.write("}\n\nfunction cleanup_gopath() {\n")

            for pkg, vers in package.iteritems():
                if 1 == len(vers):
                    cleandep.write("\trm -rf {}\n".format(pkg))

            cleandep.write("}\n")