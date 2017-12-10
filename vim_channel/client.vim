let s:delimiter = fnamemodify(".", ":p")[-1:]
let s:ext = ""
if has('win32')
  let s:ext = ".exe"
endif

let BusServer = expand("<sfile>:p:h") . s:delimiter . "server" .s:ext
command! BusInit let BusJob = job_start(BusServer)
command! BusQuit call job_stop(BusJob)
command! Bus call BusQuery()

function! BusAnnounce(handle, message)
  echo join(["次のバスは", a:message, "に到着します．"], " ")
  call ch_close(a:handle)
endfunction
function! BusQuery()
  let s:handle = ch_open("localhost:6868")
  call ch_sendexpr(s:handle, strftime("%H:%M"), {"callback": "BusAnnounce"})
endfunction
