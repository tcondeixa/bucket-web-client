function setProperties(pageNumber) {
    let bucket = document.getElementById("bucket").value;
    let filesPage = document.getElementById("filesPage").value;
    let filesOrder = document.getElementById("filesOrder").value;

    let newUrl = '/main/' + bucket + "?filesPage=" + filesPage + "&page=" + pageNumber + "&orderObjects=" + filesOrder
    console.log(newUrl);
    window.location = newUrl;
}