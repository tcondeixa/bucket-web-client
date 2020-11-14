function setBucket() {
    let bucket = document.getElementById("bucket").value;

    let newUrl = '/main/' + bucket
    console.log(newUrl);
    window.location = newUrl;
}