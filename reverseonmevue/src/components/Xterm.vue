<script>
import 'xterm/css/xterm.css'
import { Terminal } from 'xterm'
import axios from 'axios'
export default {
  name: 'Xterm',
  props: {
    host: {
      type: String,
      default: window.location.host
    },
  },
  data() {
    return {
      wsid: null,
      username: null,
      port: null,
      buffer: "",
    }
  },
  mounted() {
  },
  beforeDestroy() {
    this.socket.close()
    this.term.dispose()
  },
  methods: {
    createListener(){
      console.log(this.$props.host)
      axios.get("http://"+this.$props.host+"/create").then((res)=>{
        console.log(res.data.ID)
        this.wsid=res.data.ID
      })
    },
    initTerm() {
      const term = new Terminal({
        cursorBlink: true,
        cursorStyle: 'underline',
      });
      document.getElementById('xterm').innerHTML=""
      term.open(document.getElementById('xterm'));
      term.focus();
      term.attachCustomKeyEventHandler((e) => {
        if(e.key=="Backspace" && e.type=="keydown"){
          this.buffer=this.buffer.slice(0,-1)
          this.term.write('\b \b')
          console.log(buffer)
          return false
        }
        if(e.key=="Backspace" && e.type=="keyup"){
          return false
        }
        if(e.key=="Enter" && e.type=="keydown"){
          this.term.writeln('')
          if (this.buffer.startsWith("\r")){
            this.buffer=this.buffer.slice(1)
          }
          console.log(btoa(this.buffer+"\n"))
          this.socket.send(this.buffer+"\n")
          this.buffer=""
          return false
        }
        if(e.key=="Enter" && e.type=="keyup"){
          // this.term.writeln('')
          return false
        }
      })
      term.onData((data) => {
        console.log(data)
        this.buffer+=data
        this.term.write(data)
      })
      this.socket.onmessage = (event) => {
        var a=event.data.toString().split("\n")
        for (var i=0;i<a.length;i++){
          this.term.writeln(a[i])
        }
        console.log(event.data)
      }
      this.term = term
    },
    initSocket () {
      this.socket = new WebSocket("ws://"+this.$props.host+"/wstotcp/"+this.wsid+"?username="+this.username)
      this.openSocket()
      this.closeSocket()
      this.errorSocket()
    },
// 打开连接
    openSocket () {
      this.socket.onopen = () => {
        this.initTerm()
      }
    },
    // 关闭连接
    closeSocket () {
      this.socket.onclose = () => {
        // console
        // this.sendData()
      }
    },
    // 连接错误
    errorSocket () {
      this.socket.onerror = () => {
        this.$message.error('websoket连接失败，请刷新！')
      }
    },
  }
}
</script>
<template>
  <div id="inputlist">
  <button @click="createListener">创建监听</button>
    <input id="wsid" v-model="wsid" placeholder="请输入WSID">
    <input id="username" v-model="username" placeholder="请输入昵称">
    <button @click="initSocket">加入shell</button>
  </div>
  <div id="xterm" class="xterm" />
</template>
