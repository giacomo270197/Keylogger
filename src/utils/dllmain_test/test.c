// A simple program that uses LoadLibrary and 
// GetProcAddress to access myPuts from Myputs.dll. 
 
#include <windows.h> 
#include <stdio.h> 
#include <unistd.h>
 
typedef int (__cdecl *MYPROC)(LPWSTR); 
 
int main( void ) 
{ 
    HINSTANCE hinstLib; 
    MYPROC ProcAdd; 
    BOOL fFreeResult, fRunTimeLinkSuccess = FALSE; 
 
    // Get a handle to the DLL module.
 
    //hinstLib = LoadLibrary(TEXT("C:\\Users\\gcaso\\Keylogger\\bin\\payloads\\calc_dll\\calcdll.dll"));
    hinstLib = LoadLibrary(TEXT("..\\..\\payloads\\calc_dll\\calcdll.dll")); 

    sleep(5); // Otherwise parent process dies before the calc shows up

    fFreeResult = FreeLibrary(hinstLib); 

    return 0;

}