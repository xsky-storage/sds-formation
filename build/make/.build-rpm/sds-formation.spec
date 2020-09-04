Name: %{_pkgname}
Version: %{_version}
Release: %{_release}%{?dist}
Summary: The distributed storage management system formation utilility
Group: Applications/Productivity

License: Commercial
Source: %{name}.tar.gz

Vendor: SDS
Packager: SDS

%global debug_package %{nil}

# is_systemd conditional
%if 0%{?fedora} >= 21 || 0%{?centos} >= 7 || 0%{?rhel} >= 7
%global is_systemd 1
%endif

# required packages for build
# most are already in the container (see contrib/builder/rpm/generate.sh)
# only require systemd on those systems
%if 0%{?is_systemd}
BuildRequires: pkgconfig(systemd)
Requires: systemd-units
%else
Requires(post): chkconfig
Requires(preun): chkconfig
# This is for /sbin/service
Requires(preun): initscripts
%endif

# required packages on install
Requires: /bin/sh
# conflicting packages

%description
sds-formation is the tool used to create resources in sds according with a template.

%prep
%if 0%{?centos} <= 6
%setup -n %{_dirname}
%else
%autosetup -n %{_dirname}
%endif

%build
./build/make.sh binary

%check
readelf -h ./bundles/%{_origversion}/binary/sds-formation

%install
# install binary
install -d $RPM_BUILD_ROOT/%{_bindir}
install -p -m 755 bundles/%{_origversion}/binary/sds-formation-%{_origversion} $RPM_BUILD_ROOT/%{_bindir}/sds-formation

# list files owned by the package here
%files
/%{_bindir}/sds-formation

%doc
# /%{_mandir}/man1/*
# /%{_mandir}/man5/*
