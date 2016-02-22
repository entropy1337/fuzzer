#include "ifuzz.h"

/*
** just verifies if the "file" exists.
** doesnt enforce it to be a regular file,
** could be a directory or fifo or symlink or whatever
*/
//文件是否存在，不存在则创建文件
void
verify_file ()
{
  struct stat statbuf;

  if (stat (VALID_FILE, &statbuf))
    {
      creat (VALID_FILE, S_IRWXU | S_IRGRP | S_IXGRP | S_IROTH | S_IXOTH);	/* 755 */
      
      //debug 
      //printf("create file=%s\n", VALID_FILE);
  }

   
  return;
}


//删除文件 
/* fairly blindly remove VALID_FILE, unless its a symlink */
void
remove_file ()
{
  struct stat statbuf;
  fprintf (stderr, "called\n");
  if (!(stat (VALID_FILE, &statbuf)))
    {
      if (!(S_ISLNK (statbuf.st_mode)))
	{
	  unlink (VALID_FILE);
    
    //debug 
    //printf("remove file=%s\n", VALID_FILE);
	  fprintf (stderr, "remove\n");
	}
    }


  return;
}
