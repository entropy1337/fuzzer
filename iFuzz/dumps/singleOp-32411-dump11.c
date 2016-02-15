#include <stdio.h>
#include <string.h>
#include <unistd.h>
int main(int argc,char *argv[])
{
/*
Mon Feb 15 11:01:52 2016
Single option: -f
*/
char *args[] = 
	{ 
	 "./testcase/singleOption//singleOption",
	 "-f",
	 "/tmp",
	 NULL
	};

char *envp[] = 
	{ 
	 "LANG=en_CN.UTF-8",
	 "DISPLAY=:0.0",
	 "PWD=/root/Desktop/Parallels Shared Folders/Home/Desktop/git/fuzzer/iFuzz",
	 "LOGNAME=root",
	 "GNOME_KEYRING_PID=3450",
	 "XAUTHORITY=/var/run/gdm3/auth-for-root-na5fGw/database",
	 "GTK_IM_MODULE=ibus",
	 "COLORTERM=gnome-terminal",
	 "DESKTOP_SESSION=default",
	 "TEXTDOMAIN=im-config",
	 "GDMSESSION=default",
	 "GNOME_KEYRING_CONTROL=/root/.cache/keyring-xAlERr",
	 "USERNAME=root",
	 "GNOME_DESKTOP_SESSION_ID=this-is-deprecated",
	 "WINDOWPATH=7",
	 "TEXTDOMAINDIR=/usr/share/locale/",
	 "DBUS_SESSION_BUS_ADDRESS=unix:abstract=/tmp/dbus-9x8I3kxrfo,guid=1d080ef93e192e83b0b04be1569f4c57",
	 "CLUTTER_IM_MODULE=ibus",
	 "XDG_DATA_DIRS=/usr/share/gnome:/usr/local/share/:/usr/share/",
	 "QT4_IM_MODULE=ibus",
	 "XDG_SESSION_COOKIE=5c41c39e7dd18b72a5591aaf52b3cdfd-1453280343.275258-2033335323",
	 "GDM_LANG=C",
	 "SHELL=/bin/zsh",
	 "WINDOWID=16777220",
	 "SSH_AGENT_PID=3548",
	 "SESSION_MANAGER=local/kali:@/tmp/.ICE-unix/3468,unix/kali:/tmp/.ICE-unix/3468",
	 "SSH_AUTH_SOCK=/root/.cache/keyring-xAlERr/ssh",
	 "TERM=xterm",
	 "PATH=/root/.autojump/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
	 "HOME=/root",
	 "USER=root",
	 "XMODIFIERS=@im=ibus",
	 "GPG_AGENT_INFO=/root/.cache/keyring-xAlERr/gpg:0:1",
	 "SHLVL=1",
	 "OLDPWD=/root/Desktop/Parallels Shared Folders/Home/Desktop/git/fuzzer/iFuzz/testcase",
	 "ZSH=/root/.oh-my-zsh",
	 "DEFAULT_USER=root",
	 "GREP_OPTIONS=--color=auto --exclude-dir=.cvs --exclude-dir=.git --exclude-dir=.hg --exclude-dir=.svn",
	 "GREP_COLOR=1;32",
	 "PAGER=less",
	 "LESS=-R",
	 "LC_CTYPE=C",
	 "LSCOLORS=Gxfxcxdxbxegedabagacad",
	 "AUTOJUMP_ERROR_PATH=/root/.local/share/autojump/errors.log",
	 "LC_ALL=en_CN.UTF-8",
	 "_=/root/Desktop/Parallels Shared Folders/Home/Desktop/git/fuzzer/iFuzz/./ifuzz",
	 NULL
	};

execve("./testcase/singleOption//singleOption",args,envp);
return 0;
}
