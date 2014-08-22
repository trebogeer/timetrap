var insertLinebreaks = function (d) {
  var el = d3.select(this);
  var words = el.text().split(' ');

  el.text('');

  for (var i = 0; i < words.length; i++) {
        var tspan = el.append('tspan').text(words[i]);
        if (i > 0)
            tspan.attr('x', 0).attr('dy', '15');
  }
};

var interpolateChart = function (chart, d3) {
     var itype = d3.event.target.value;
     chart = chart.interpolate(itype);
     var svg = d3.select('#chart svg');
     svg.call(chart);
     nv.utils.windowResize(chart.update);
}


var Chart = {
  chart: null,
  d3: null,
  nv: null,
  svg: null,
  alias: null,
  init: function (d3, nv, queryString) {

    var chart;
    var ctype = queryString["t"];
    var intGuide = queryString["ig"];
    if(ctype != null && "cch" == ctype) {
         chart = nv.models.cumulativeLineChart().x(function(d) { return d[0] }).y(function(d) { return d[1] });
    } else if (ctype != null && "lfch" == ctype) {
        chart = nv.models.lineWithFocusChart().x(function(d) { return d[0] }).y(function(d) { return d[1] });
        chart.x2Axis
        .tickFormat(function(d) {
            return d3.time.format('%X')(new Date(d));
         });
    } else {
        chart = nv.models.lineChart().x(function(d) { return d[0] }).y(function(d) { return d[1] });
    }

    if(intGuide != null && "" != intGuide && ctype != "lfch") {
        chart = chart.useInteractiveGuideline(true);
    }

    var interpol = queryString["interpolate"];
    if(interpol != null && "" != interpol) {
        chart = chart.interpolate(interpol);
    }

    chart.xAxis
    .tickFormat(function(d) {
       return d3.time.format('%X %x')(new Date(d));
    });
    Chart.chart = chart;
    Chart.d3 = d3;
    Chart.nv = nv;
    Chart.svg = d3.select('#chart svg');
  },
  interpolate:function() {
    var itype = Chart.d3.event.target.value;
    Chart.chart = Chart.chart.interpolate(itype);
    Chart.svg.call(Chart.chart);
    Chart.d3.selectAll('.nv-axisMaxMin text').each(insertLinebreaks);
    Chart.d3.selectAll('.nvd3.nv-wrap.nv-axis g g text').each(insertLinebreaks);
    Chart.nv.utils.windowResize(Chart.chart.update);
  },
  render:function(dataLink) {
    var data;
    Chart.d3.json(dataLink,
        function(err, d) {
           if(err) {return console.error(err)}
           data = d['data'];
           Chart.alias = d['alias']
           var svg = d3.select('#chart svg');
           Chart.svg.datum(data)
           Chart.svg.call(Chart.chart);
           Chart.d3.select('#gtitle').text(Chart.alias);
           Chart.d3.selectAll('.nv-axisMaxMin text').each(insertLinebreaks);
           Chart.d3.selectAll('.nvd3.nv-wrap.nv-axis g g text').each(insertLinebreaks);
        });

     Chart.nv.utils.windowResize(Chart.chart.update);
  }

}