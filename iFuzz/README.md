##iFuzz
iFuzz是一个本地模糊测试器，可以自动处理各种目标二进制代码，生成C语言触发器，重现缺陷。


##代码
原版： [代码](http://fuzzing.org/wp-content/ifuzz.tar)



##模块介绍
- argv[0]模糊测试
- argv[1]模糊测试
- 多选项模糊测试
- getopt模糊测试（getopt钩子） 
- getenv模糊测试


##使用方法
	ifuzz <fuzztype> <binary directory> [fuzz specific options]

	Fuzztypes:  
	0 - argv[0] fuzzing
	1 - argv[1] fuzzing
	2 - incremental single option fuzzing
	3 - incremental multiple option fuzzing

	Example:
	ifuzz 3 directory/ <-o optstring> [-e extra-args] [-f first_arg] [-l last_arg] [-s]
	ifuzz 1 directory/ [-s]
	ifuzz 0 directory/ [-s]


##新增功能（未完成）
- 使用特定的字符串数据库(文件读取)
- 自动生成指定长度的随机字符串
- 自动搜索系统中setuid和setgid的程序
- 优化自动生成C代码及自动编译重现，自动报告等


