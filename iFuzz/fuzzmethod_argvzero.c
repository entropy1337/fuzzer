/* $Id */

#include "ifuzz.h"

extern pid_t child;

/*
** fullpath: path binaries reside in
*/

//argv0测试函数
void
fuzzmethod_argvzero (char *fullpath, struct argv_args *argv_args)
{
  pid_t pid;
  int status;

  char *args[6];
  FILE *fp;
  
  //随机args[0] 
  args[0] = get_random_string();

  args[1] = "-h";
  args[2] = "-z";
  args[3] = "-zz";
  args[4] = "----";
  args[5] = NULL;


   //创建进程
  if ((pid = fork ()) != 0)
    {
      child = pid;
      signal (SIGALRM, &handle_alarm);
      alarm (TIME_TO_DIE);
      waitpid (pid, &status, 0);
      alarm (0);

      //判断进程状态
      if (WIFSIGNALED (status))
	{

		//处理信号
	  switch (WTERMSIG (status))
	    {
	      /* 
	       ** since we are only logging the signals, we might as well catch anything
	       ** even remotely interesting
	       */
	    case SIGBUS:
	    case SIGILL:
	    case SIGSEGV:
	    case SIGTRAP:
	    case SIGFPE:
	    case SIGUSR1:
	    case SIGUSR2:
//            fprintf (stderr, "%s | CRASH SIGNAL #%d (argv[0]) ", fullpath,
//                     WTERMSIG (status));
	      fprintf (stderr, "CRASH\n");
	      if (!(fp = open_c_file (fullpath, pid, WTERMSIG (status))))

		{
		  fprintf (stderr,
			   "have you ever heard of chmod?  no access to dump dir you douchebag.\n");
		  exit (-1);
		}
		  //生成重放的.c文件
	      print_c_basic_header (fp);
	      print_c_comment_open (fp);
	      print_text (fp, asciitime ());
	      print_text (fp, "Standard argv[0] crash");
	      print_c_comment_close (fp);
	      print_c_array_to_file (fp, args, "args");
	      print_c_array_to_file (fp, environ, "envp");
	      print_c_execve_call (fp, fullpath, "args", "envp");
	      print_c_basic_header_close (fp);
	      fclose (fp);

	      break;
	    default:
	      break;
	    }
	}

    }
  else
    {
      /* do the actual fuzz */
      if (argv_args->silent)	/* silent mode */
	{
	  int fd;
	  fd = open ("/dev/null", O_WRONLY);

	  //重定向输出和错误
	  dup2 (fd, STDOUT_FILENO);
	  dup2 (fd, STDERR_FILENO);
	}

	  //执行  使用不定长参数
      execle (fullpath, get_random_string(), "-h", "-z", "-zz", "----", NULL,
	      environ);

      perror ("execle");

    }
  rfree ();
  return;			/* unreached */
}
