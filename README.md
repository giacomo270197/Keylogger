# Malware_Techniques_Implementations

A little project to practice techniques shown in Practical Malware Analysis. I believe the best way to spot malware behavior is to know exactly what it takes to actually program it. I will try to avoid implementing anyhting that can be used out-of-the-box for any malicious intent.
The project will be mainly written in Go, for several reasons. I like the static linking properties, and it will be interesting to see how Defender detectes some of these techniques. Also, I expect tons of sample C/C++ code out there for the things I will implement, so writing it in Go should prove a bit more of a challenge. Finally, I just wanna be as comfortable with Go as I am with C.

## Launching Techniques

Following are the malware lauching techniques I implemented.

### DLL Injection

DLL injection is implemented under `src/injectors/dll_injection`. Some sample DLL are also created. They spawn either a Command Prompt or a Calculator. This is not really useful from an attacking standpoint, the injected DLLs should not  spawn processes. This is just done as an example.
Through Process Explorer I can confirm that the process manages to get the SeDebugPrivilege and is able to inject, from an Administrator session, into a SYSTEM process.
Not all SYSTEM processes are injectables, but I cannot seem to understand which ones are and which ones are not.



