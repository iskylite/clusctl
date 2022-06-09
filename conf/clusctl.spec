%define name clusctl
%define version 1.6.1
%define release 2

Summary: %{name} - HPC Cluster Control And Manage Tools
Name: %{name}
Version: %{version}
Release: %{release}
Source0: %{name}-%{version}.tar.gz
License: Apache-2.0
Group: Development/Libraries
BuildRoot: %{_tmppath}/%{name}-%{version}-%{release}-buildroot
Prefix: %{_prefix}
Vendor: iskylite <yantao0905@outlook.com>
URL: https://github.com/iskylite/%{name}.git

%description
# %{name} - HPC Cluster Control And Manage Tools

USAGE:
   %{name} [global options] command [command options] [arguments...]

VERSION:
   v1.6.0

AUTHOR:
   iskylite <yantao0905@outlook.com>

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --background, -b           run cmd in background (default: false)
   --command value, -c value  [action] linux shell command to run
   --debug, -D                set log level debug (default: false)
   --dest DIR, -d DIR         dest DIR on remote host (default: "/home/yanta")
   --disablecolor, --dc       disable log color print (default: false)
   --hostfile FILE, -H FILE   path to FILE containing a list of target hosts
   --list, -l                 sort command output by node list (default: false)
   --nodes value, -n value    app agent nodes list
   --output value, -o value   dump rcopy or command output into log file
   --ping, -p                 [action] check all agent status (default: false)
   --port value, -P value     grpc service port (default: 1995)
   --rcopy FILE, -r FILE      [action] local FILE path
   --size value, -s value     payload size (eg: 51200, 512k, 1m) (default: "2M")
   --timeout value, -t value  timeout for ping (default: 1)
   --workers value, -w value  multi open workers (default: 64)
   --help, -h                 show help (default: false)
   --version, -v              print the version 1.6.1


%prep

%build

%install
mkdir $RPM_BUILD_ROOT/usr/sbin -p
mkdir $RPM_BUILD_ROOT/var/lib/%{name}d -p
mkdir $RPM_BUILD_ROOT/etc/systemd/system/ -p

cp -ap %{_builddir}/%{name}-%{version}/usr/sbin/%{name} $RPM_BUILD_ROOT/usr/sbin
cp -ap %{_builddir}/%{name}-%{version}/usr/sbin/%{name}d $RPM_BUILD_ROOT/usr/sbin
cp -ap %{_builddir}/%{name}-%{version}/etc/systemd/system/%{name}d.service $RPM_BUILD_ROOT/etc/systemd/system/
cp -ap %{_builddir}/%{name}-%{version}/var/lib/%{name}d/* $RPM_BUILD_ROOT/var/lib/%{name}d

chmod u+x $RPM_BUILD_ROOT/usr/sbin/*

%clean
rm -rf $RPM_BUILD_ROOT

%post
systemctl daemon-reload
systemctl enable --now %{name}d.service

%preun
systemctl disable --now  %{name}d.service

%postun
rm -rf /etc/systemd/system/%{name}d.service
rm -rf /var/lib/clusctld

%files
%defattr(-,root,root)
/etc/systemd/system/%{name}d.service
/usr/sbin/%{name}
/usr/sbin/%{name}d
/var/lib/%{name}d/cert.key
/var/lib/%{name}d/cert.pem
