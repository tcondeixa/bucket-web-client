function setProperties(pageNumber) {
    let bucket = document.getElementById("bucket").value;
    let filesPage = document.getElementById("filesPage").value;

    let newUrl = '/main/' + bucket + "?filesPage=" + filesPage + "&page=" + pageNumber
    console.log(newUrl);
    window.location = newUrl;
}