const socket = await Bun.connect({
    hostname: "localhost",
    port: 3000,
    socket: {
        open(socket){
            console.log('hi')
            socket.write(JSON.stringify({name: "foo", key: "foo"}) + '\n')
        },
        data(socket, buffer){
            console.log(buffer.toString('utf-8'))
        },
        error(_, error){
            console.error(error)
        },
        close(){
            console.warn('closed')
        }
    }
})

export {}