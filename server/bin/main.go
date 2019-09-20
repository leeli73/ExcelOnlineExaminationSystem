package main

import (
    "fmt"
    "time"
    "syscall"
    "unsafe"
	"os/exec"
	//_ "github.com/icattlecoder/godaemon"
	registry "github.com/golang/sys/windows/registry"
)

type ulong int32
type ulong_ptr uintptr

type PROCESSENTRY32 struct {
    dwSize ulong
    cntUsage ulong
    th32ProcessID ulong
    th32DefaultHeapID ulong_ptr
    th32ModuleID ulong
    cntThreads ulong
    th32ParentProcessID ulong
    pcPriClassBase ulong
    dwFlags ulong
    szExeFile [260]byte
}

func main() {
    MonitorProcess()
    go MonitorUDisk()
}

func MonitorUDisk(){
	for ;;{
        k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Services\USBSTOR\Enum`, registry.QUERY_VALUE)
        if err != nil {
            fmt.Println(err)
        }
        defer k.Close()

        n, _, err := k.GetIntegerValue("Count")
        if err != nil {
            fmt.Println(err)
        }
        if n < 1 {
            fmt.Println("没有检测到u盘！")
        } else {
            fmt.Println("检测到U盘插入！")
        }
        time.Sleep(5 * time.Second)
    }
}

func MonitorProcess(){
    var list = [4]string{"QQ.exe","TIM.exe","chrome.exe","wechat.exe"}
    for ;;{
        kernel32 := syscall.NewLazyDLL("kernel32.dll");
        CreateToolhelp32Snapshot := kernel32.NewProc("CreateToolhelp32Snapshot");
        pHandle,_,_ := CreateToolhelp32Snapshot.Call(uintptr(0x2),uintptr(0x0));
        if int(pHandle)==-1 {
            
        }
        Process32Next := kernel32.NewProc("Process32Next");
        for {
            var proc PROCESSENTRY32;
            proc.dwSize = ulong(unsafe.Sizeof(proc));
            if rt,_,_ := Process32Next.Call(uintptr(pHandle),uintptr(unsafe.Pointer(&proc)));int(rt)==1 {
                for _,v := range list{
                    if v == string(proc.szExeFile[0:]) {
                        fmt.Println("检测到敏感进程")
                    }
                }
            }else{
                break;
            }
        }
        CloseHandle := kernel32.NewProc("CloseHandle");
        _,_,_ = CloseHandle.Call(pHandle);
        time.Sleep(5 * time.Second)
    }
}

func ShowMsg(msg string){
	cmd := exec.Command(`mshta`,`vbscript:window.execScript("alert('hello world!');","javascript")`)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	err = cmd.Wait() 
}