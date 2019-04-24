package nsenter

/*
#define _GNU_SOURCE
#include <unistd.h>
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
__attribute__((constructor)) void enter_namespace(void) {
	char *lightc_pid;
	lightc_pid = getenv("lightc_pid");
	if (lightc_pid) {
		//fprintf(stdout, "got lightc_pid=%s\n", lightc_pid);
	} else {
		//fprintf(stdout, "missing lightc_pid env skip nsenter");
		return;
	}
	char *lightc_cmd;
	lightc_cmd = getenv("lightc_cmd");
	if (lightc_cmd) {
		//fprintf(stdout, "got lightc_cmd=%s\n", lightc_cmd);
	} else {
		//fprintf(stdout, "missing lightc_cmd env skip nsenter");
		return;
	}
	int i;
	char nspath[1024];
	char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt" };
	for (i=0; i<5; i++) {
		sprintf(nspath, "/proc/%s/ns/%s", lightc_pid, namespaces[i]);
		int fd = open(nspath, O_RDONLY);
		if (setns(fd, 0) == -1) {
			//fprintf(stderr, "setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
		} else {
			//fprintf(stdout, "setns on %s namespace succeeded\n", namespaces[i]);
		}
		close(fd);
	}
	int res = system(lightc_cmd);
	exit(0);
	return;
}
*/
import "C"
