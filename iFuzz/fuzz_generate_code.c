//重现缺陷  保留现场环境
//自动生成.c文件 

//TODO 
//自动生成makefile
//自动报告和执行
     

#include "ifuzz.h"

void
print_c_basic_header (FILE * fp)
{
  fprintf (fp, "#include <stdio.h>\n");
  fprintf (fp, "#include <string.h>\n");
  fprintf (fp, "#include <unistd.h>\n");

  fprintf (fp, "int main(int argc,char *argv[])\n");
  fprintf (fp, "{\n");

}

void
print_c_comment (FILE * fp, char *comment)
{
  print_c_comment_open (fp);
  fprintf (fp, "%s\n", comment);
  print_c_comment_close (fp);
  return;
}

void
print_text (FILE * fp, char *text)
{
  fprintf (fp, "%s", text);
  return;
}


//输出执行语句
/* args and envp are symbol names, path is an actual string for the path */
void
print_c_execve_call (FILE * fp, char *path, char *args, char *envp)
{
  //TODO path应该改为文件夹的相对路径
  fprintf (fp, "execve(\"%s\",%s,%s);\n", path, args, envp);
  return;
}



void
print_c_basic_header_close (FILE * fp)
{
  fprintf (fp, "return 0;\n");
  fprintf (fp, "}\n");
  return;
}

//写入数据   argv和 环境变量的数组
void
print_c_array_to_file (FILE * fp, char *array[], char *arrayname)
{
  int ix = 0;
  fprintf (fp, "char *%s[] = \n\t{ \n", arrayname);
  while (array[ix])
    {
      fprintf (fp, "\t \"%s\",\n", array[ix++]);
    }
  fprintf (fp, "\t NULL\n\t};\n\n");
}


//打开文件
FILE *
open_c_file (char *binary, int pid, int signal)
{
  char filename[512];
  char *ptr;
  char *result;
  ptr = result = binary;

  while ((ptr = strstr (ptr, "/")))
    {
      result = ++ptr;
    }
  ptr = result;
  if (!result)
    result = binary;
  snprintf (filename, sizeof (filename) - 1, "%s/%.8s-%05d-dump%.2d.c",
	    CODE_DUMP_PATH, result, pid, signal);
  return fopen (filename, "w");
}


//输出注释 标记时间和漏洞类型
void
print_c_comment_open (FILE * fp)
{
  fprintf (fp, "/*\n");
  return;
}

void
print_c_comment_close (FILE * fp)
{
  fprintf (fp, "\n*/\n");
  return;
}
