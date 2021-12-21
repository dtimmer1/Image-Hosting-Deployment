window.addEventListener('load', function(){
    document.querySelector('input[type="file"]').addEventListener('change', function(){
        if(this.files && this.files[0]){
            var file = this.files[0]
            var img = document.querySelector('img')
            img.onload = () => {
                URL.revokeObjectURL(img.src)
            }

            var filename = URL.createObjectURL(file)

            img.src = filename
            img.style = "display: block;"

            var reader = new FileReader();

            reader.readAsDataURL(file);

            reader.onload = function(){
                var httpPost = new XMLHttpRequest(),
                path = "http://localhost:8080/send",
                data = JSON.stringify({image: reader.result});
            
                httpPost.onreadystatechange = function(err) {
                    if (httpPost.readyState == 4 && httpPost.status == 200){
                        console.log(httpPost.responseText);
                        var url = `http://localhost:8080/img/${httpPost.responseText}`
                        document.getElementById("link").href = url
                        document.getElementById("thankyou").style = "display: block;"

                    } else {
                        console.log(err);
                    }
                }

                httpPost.open("POST", path, true)
                httpPost.setRequestHeader('Content-Type', 'application/json')
                httpPost.setRequestHeader('Access-Control-Allow-Origin', '*')
                httpPost.send(data)
            };
        }
    });
});