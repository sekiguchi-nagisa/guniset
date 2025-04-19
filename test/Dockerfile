FROM opensuse/tumbleweed

RUN zypper addrepo --no-gpgcheck -f https://download.opensuse.org/repositories/home:nsekiguchi/openSUSE_Tumbleweed/home:nsekiguchi.repo && \
    zypper refresh && \
    zypper install -y arsh git go diffutils

COPY . /home/tux/guniset

# under Github Actions, regardless of WORKDIR setting, WORKDIR always indicates GITHUB_WORKSPACE (source code location)
# so, if create directory at WORKDIR, need root privilege
# (https://docs.github.com/en/actions/creating-actions/dockerfile-support-for-github-actions)
WORKDIR /home/tux/guniset/

CMD DIR="$(pwd)" && cd  /home/tux/guniset && git config --global --add safe.directory "${PWD}" && \
    ./scripts/build.sh && ./scripts/get_unicode_data.sh && \
    GUNISET_DIR=./unicode_dir ./scripts/run_e2e_test.arsh ./guniset
