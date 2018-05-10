

var numbers = [4, 9, 16, 17];

var n = 0;
numbers.forEach(function(item){
    var a = 0;
    if(n >0){
        a = item - n
        console.log(a);
    }
    n = item;
})