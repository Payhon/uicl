// Interface fixture only. This intentionally refuses to modify the operating system.
using System;
public static class ServiceActions
{
    public static void SetStartMode(string service, string mode)
    {
        throw new NotSupportedException("Bind a reviewed, allowlisted broker implementation before execution.");
    }
}
